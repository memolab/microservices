FROM alpine:3.23
WORKDIR /app
COPY build/api_gateway_service .
ENTRYPOINT [ "/app/api_gateway_service" ]
