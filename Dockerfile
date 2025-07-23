FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o app main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/app /app/app
COPY web /app/web
COPY config.yaml /app/config.yaml
COPY prometheus.yml /app/prometheus.yml
EXPOSE 8080 8081 8082
CMD ["/app/app"] 