package main

import (
	"time"

	"microservices-demo/common/env"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// var clientCallOpts = []grpc.CallOption{grpc.WaitForReady(true)}

// newUserStoreClient return new grpc Client, MUST Close()
func newUserStoreClient() (*grpc.ClientConn, error) {
	addr := "dns:///" + env.GetString("USER_STORE_SERVICE_NAME", "") + ":" + env.GetString("USER_STORE_SERVICE_PORT", "")

	conn, err := grpc.NewClient(addr, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                60 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	}...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
