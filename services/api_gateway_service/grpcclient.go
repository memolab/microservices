package main

import (
	"microservices-demo/common/env"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//var clientCallOpts = []grpc.CallOption{grpc.WaitForReady(true)}

// newUserStoreClient return new grpc Client, MUST Close()
func newUserStoreClient() (*grpc.ClientConn, error) {
	addr := env.GetString("USER_STORE_SERVICE_NAME", "") + ":" + env.GetString("USER_STORE_SERVICE_PORT", "")
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	}
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
