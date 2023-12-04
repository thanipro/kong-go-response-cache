# Build stage
FROM golang:latest as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o response-cacher

FROM kong:2.8.1-alpine

COPY --from=builder /app/response-cacher /usr/local/bin/
USER root
RUN chmod a+x /usr/local/bin/response-cacher
USER kong
ENTRYPOINT ["/usr/local/bin/response-cacher"]
EXPOSE 8000 8443 8001 8444
STOPSIGNAL SIGQUIT
HEALTHCHECK --interval=10s --timeout=10s --retries=10 CMD kong health
CMD ["kong", "docker-start"]
