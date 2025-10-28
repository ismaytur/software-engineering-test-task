FROM alpine:latest

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY bin/server ./server
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh

RUN chmod +x /app/server /usr/local/bin/entrypoint.sh

ENV POSTGRES_HOST=db \
    POSTGRES_PORT=5432 \
    POSTGRES_USER=postgres \
    POSTGRES_PASSWORD=postgres \
    POSTGRES_DB=postgres \
    POSTGRES_SSL_MODE=disable

EXPOSE 8080

USER app

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
