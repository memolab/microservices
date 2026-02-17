FROM alpine:3.23
WORKDIR /app
COPY build/user_store_service .
ENTRYPOINT [ "/app/user_store_service" ]
