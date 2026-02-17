package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"microservices-demo/common/env"
)

func init() {
	slogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.Level(env.GetInt("LOG_LEVEL", -4)), AddSource: true, ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				source.File = filepath.Base(source.File)
			}
		}
		return a
	}}))
	slog.SetDefault(slogger)
}

func main() {
	sctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup
	defer func() {
		stop()
		wg.Wait()
		slog.Info("api-gateway-service stopped")
	}()

	slog.Info("api-gateway-service start", "environment", env.GetString("ENVIRONMENT", "dev"))
	wg.Go(func() {
		start(sctx, stop)
	})

	defer slog.Info("api-gateway-service shutting down...")

	<-sctx.Done()
}
