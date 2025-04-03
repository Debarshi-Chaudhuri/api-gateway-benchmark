FROM golang:1.23.7-alpine AS builder

WORKDIR /app

# Copy go.mod from root
COPY go.mod ./
# Run go mod tidy to ensure go.sum is up to date
RUN go mod tidy
RUN go mod download

# Copy the HTTP service code
COPY http-service/main.go ./

# Build the application
RUN go build -o http-service main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl openssl

WORKDIR /app/

COPY --from=builder /app/http-service .

# Create directory for TLS certificates
RUN mkdir -p /etc/http-service/certs

# Copy TLS certificates
COPY tls/http/ /etc/http-service/certs/

EXPOSE 8000

CMD ["./http-service"]

HEALTHCHECK --interval=5s --timeout=3s --retries=3 \
  CMD curl -f --cacert /etc/http-service/certs/ca.crt https://localhost:8000/health || exit 1