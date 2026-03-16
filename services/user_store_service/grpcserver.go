package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"microservices-demo/common/env"
	"microservices-demo/common/messaging"
	"microservices-demo/common/observe"
	"microservices-demo/common/pb/v1"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type UserService struct {
	pb.UnimplementedUserServiceServer
	messageConsumer *userConsumer
}

func start(sctx context.Context, stop context.CancelFunc) {
	// init RabbitMq
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		slog.Error("failed to connect rabbitmq", "error", err)
		stop()
		return
	}
	defer rabbitmq.Close()

	// init Jaeger
	shutdownTracer, err := observe.InitGlobalTracer(env.GetString("USER_STORE_SERVICE_NAME", ""), env.GetString("JAEGER_ENDPOINT", ""))
	if err != nil {
		slog.Error("failed to init tracer", "error", err)
		stop()
		return
	}

	// init Messageing Consumer
	messageConsumer := newMessageConsumer(rabbitmq)
	if err := messageConsumer.Listen(sctx); err != nil {
		slog.Error("fail listening user consumer", "error", err)
		stop()
		return
	}

	// init Prometheus
	prometheusMetrics := observe.InitPromMetrics()
	prometheusHTTP := observe.InitPrometheusHTTPServer(env.GetString("METRICS_PORT", ""))

	grpcServer := grpc.NewServer([]grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute,
			MaxConnectionAge:      30 * time.Minute,
			MaxConnectionAgeGrace: 10 * time.Second,
			Time:                  2 * time.Minute,
			Timeout:               20 * time.Second,
		}),
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
		grpc.ChainUnaryInterceptor(
			prometheusMetrics.UnaryServerInterceptor(),
		),
	}...)
	pb.RegisterUserServiceServer(grpcServer, &UserService{messageConsumer: messageConsumer})

	prometheusMetrics.InitializeMetrics(grpcServer)

	addr := ":" + env.GetString("USER_STORE_SERVICE_PORT", "")
	l, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("grpc net listen fail", "error", err)
		stop()
		return
	}

	go func() {
		<-sctx.Done()
		tctx, cancel := context.WithTimeout(sctx, 10*time.Second)
		defer cancel()
		grpcServer.GracefulStop()
		shutdownTracer(tctx)
		prometheusHTTP.Shutdown(tctx)
		l.Close()
		go func() {
			<-tctx.Done()
			grpcServer.Stop()
		}()
	}()

	go func() {
		if err := prometheusHTTP.ListenAndServe(); err != nil {
			slog.Error("prometheus http metrics listen fail", "error", err)
			stop()
			return
		}
	}()

	slog.Info("grpc server start", "addr", addr)
	if err := grpcServer.Serve(l); err != nil {
		if !errors.Is(err, grpc.ErrServerStopped) {
			slog.Error("grpc serve fail", "error", err, "addr", addr)
			stop()
		}
	}
	slog.Info("grpc server closed", "addr", addr)
}
