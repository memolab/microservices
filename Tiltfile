# Load the restart_process extension
load('ext://restart_process', 'docker_build_with_restart')

k8s_yaml('./app-deploy/dev/k8s/app-config.yaml')
k8s_yaml('./app-deploy/dev/k8s/secrets.yaml')

### rabbitMQ
k8s_yaml('./app-deploy/dev/k8s/rabbitmq.yaml')
k8s_resource('rabbitmq', port_forwards=['5672', '15672'], labels="tooling")

### endpoint: /api-gateway
local_resource('api-gateway-compile', "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api_gateway_service ./services/api_gateway_service",
deps=['./services/api_gateway_service'], trigger_mode=TRIGGER_MODE_MANUAL, labels="compiles")
docker_build_with_restart(
  'microservices/api-gateway-service',
  '.',
  entrypoint=['/app/api_gateway_service'],
  dockerfile='./app-deploy/dev/docker/api-gateway-service.Dockerfile',
  only=[
    './build/api_gateway_service',
  ],
  live_update=[
    sync('./build/api_gateway_service', '/app/api_gateway_service'),
  ],
)
k8s_yaml('./app-deploy/dev/k8s/api-gateway.yaml')
k8s_resource('api-gateway-service', port_forwards=8080, resource_deps=['api-gateway-compile'], labels="services")

### endpoint: /user-store-service
local_resource('user-store-compile', "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/user_store_service ./services/user_store_service",
deps=['./services/user_store_service'], trigger_mode=TRIGGER_MODE_MANUAL, labels="compiles")
docker_build_with_restart(
  'microservices/user-store-service',
  '.',
  entrypoint=['/app/user_store_service'],
  dockerfile='./app-deploy/dev/docker/user-store-service.Dockerfile',
  only=[
    './build/user_store_service',
  ],
  live_update=[
    sync('./build/user_store_service', '/app/user_store_service'),
  ],
)
k8s_yaml('./app-deploy/dev/k8s/user-store.yaml')
k8s_resource('user-store-service', port_forwards=7890, resource_deps=['user-store-compile'], labels="services")

### endpoint: /jaeger
k8s_yaml('./app-deploy/dev/k8s/jaeger.yaml')
k8s_resource('jaeger', port_forwards=['16686:16686'], labels="tooling")
