# Microservices (Go)

A small Go-based microservices example providing a gRPC user service and an API gateway. This repository includes local development tooling, Dockerfiles, and Kubernetes manifests for dev/prod deployments.

## Repository layout

- `services/` — service implementations
  - `api_gateway_service/` — API gateway and HTTP handlers
  - `user_store_service/` — user store gRPC service
- `common/` — shared utilities and generated protobuf code
- `proto/` — source .proto files (`proto/v1/user_service.proto`)
- `app-deploy/` — deployment artifacts (Dockerfiles, k8s manifests) for `dev` and `prod`, Note: removed secrets.yaml file
- `build/` — pre-built service binaries (optional)

## Requirements

- Go 1.25+ (or compatible)
- Docker (for building images)
- kubectl (for applying k8s manifests)
- Optional: Tilt, buf/protoc toolchain for protobuf generation

## Quickstart — local (simple)

### Tilt (local development)

This repository includes a `Tiltfile` at the repo root for rapid local development. Tilt will build images, apply Kubernetes manifests, and stream logs for the services in `app-deploy/dev/k8s/`.

Quick start with Tilt:

```bash
# Install Tilt: https://tilt.dev
tilt up
```

Tilt will read the `Tiltfile` and bring up the development environment. Use the Tilt UI at `http://localhost:10350` to view logs, resources, and to restart services.

If you prefer to run a single service with Tilt, use the Tilt UI or configure the `Tiltfile` to expose individual service targets.

## Protobufs / gRPC

Proto sources are in `proto/v1/`. Generated Go code is stored under `common/pb/v1/` (check `make:genproto`). To regenerate protobufs you can use `protoc` or `buf` depending on your setup. Example using `protoc` (requires plugins):

```bash
make genproto
```

## Observability

The repository includes tracing and metrics helpers under `common/observe/` (Jaeger, Prometheus integration). Deploy monitor/observability manifests from `app-deploy/dev/k8s/` as needed.
