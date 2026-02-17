FROM golang:1.25.6-alpine3.23 as builder
WORKDIR /app
COPY common common
COPY go.mod go.sum ./
RUN go mod tidy
COPY services/user_store_service services/user_store_service
RUN go build -trimpath -ldflags "-s -w" -o build/user_store_service ./services/user_store_service


FROM alpine:3.23
WORKDIR /app
COPY --from=builder app/build/user_store_service .
ENTRYPOINT [ "/app/user_store_service" ]
