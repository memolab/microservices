FROM golang:1.25.6-alpine3.23 as builder
WORKDIR /app
COPY common common
COPY go.mod go.sum ./
RUN go mod tidy
COPY services/api_gateway_service services/api_gateway_service
RUN go build -trimpath -ldflags "-s -w" -o build/api_gateway_service ./services/api_gateway_service


FROM alpine:3.23
WORKDIR /app
COPY --from=builder app/build/api_gateway_service .
ENTRYPOINT [ "/app/api_gateway_service" ]
