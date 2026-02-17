package main

import (
	"context"
	"embed"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"microservices-demo/common/env"
	"microservices-demo/common/messaging"
	"microservices-demo/common/observe"
	"microservices-demo/common/rid"
	"microservices-demo/common/types"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

//go:embed static/*
var fsAssets embed.FS

type cntrlHandlers struct {
	messagePublisher *messagePublisher
}

func start(sctx context.Context, stop context.CancelFunc) {
	rabbitMqURI := env.GetString("RABBITMQ_URI", "")
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		slog.Error("failed to init rabbitmq", "error", err)
		stop()
		return
	}
	defer rabbitmq.Close()

	shutdownTracer, err := observe.InitGlobalTracer(env.GetString("API_GATEWAY_SERVICE_NAME", ""), env.GetString("JAEGER_ENDPOINT", ""))
	if err != nil {
		slog.Error("failed to init tracer", "error", err)
		stop()
		return
	}

	handlers := &cntrlHandlers{messagePublisher: newMessagePublisher(rabbitmq)}
	mux := http.NewServeMux()

	// static files handler
	muxFiles := http.NewServeMux()
	muxFiles.Handle("/", loggingMiddleware(setAssetsHeadersMiddleware(http.FileServerFS(fsAssets), time.Hour)))
	mux.Handle("/static/", muxFiles)

	// api handlers
	muxPages := http.NewServeMux()
	muxPages.HandleFunc("GET /{$}", handlers.index)
	// set user to user-store-service via rabbitmq
	muxPages.Handle("GET /setuser/{id}", otelhttp.NewHandler(http.HandlerFunc(handlers.setUser), "/setuser/{id}"))
	// get user back from user-store-service via grpc client
	muxPages.Handle("GET /getuser/{id}", otelhttp.NewHandler(http.HandlerFunc(handlers.getUser), "/getuser/{id}"))

	handlersStack := applyMiddlewares(loggingMiddleware, setHeadersMiddleware)
	// otelhttp.NewMiddleware("request-http", otelhttp.WithMetricAttributesFn(func(r *http.Request) []attribute.KeyValue {
	// 	span := trace.SpanFromContext(r.Context())
	// 	cid := string(r.Context().Value(types.RequestIDKey{}).(string))
	// 	span.SetAttributes(attribute.String("request-id", cid))
	// 	//return []attribute.KeyValue{attribute.String("request-id", rid)}
	// 	return []attribute.KeyValue{attribute.String("request", "http")}
	// })),

	mux.Handle("/", handlersStack(muxPages))

	addr := ":" + env.GetString("API_GATEWAY_SERVICE_PORT", "")
	srvr := &http.Server{
		Addr:     addr,
		Handler:  mux,
		ErrorLog: slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(context.Background(), types.ConnIDKey{}, rid.New8())
		},
	}

	go func() {
		<-sctx.Done()
		tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srvr.Shutdown(tctx); err != nil {
			slog.Error("shutdown http")
		}
		shutdownTracer(tctx)
	}()

	slog.Info("http server start", "addr", addr)
	if err := srvr.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server fail", "error", err, "addr", addr)
			stop()
		}
	}
	slog.Info("http server closed", "addr", addr)
}
