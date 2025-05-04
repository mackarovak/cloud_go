FROM golang:1.24.2-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o load-balancer .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/load-balancer .
COPY config/config.yaml ./config/

EXPOSE 8080
CMD ["./load-balancer"]