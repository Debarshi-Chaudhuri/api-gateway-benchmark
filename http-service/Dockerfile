FROM golang:1.23.7 AS build-image

WORKDIR /go/src/app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY http-service/ ./http-service/

# Build the binary with explicit architecture settings
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/http-service http-service/main.go

# Create a minimal runtime image
FROM alpine:latest

# Add CA certificates and basic tools
RUN apk add --no-cache ca-certificates curl tzdata

# Create a non-root user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the binary from the build image
COPY --from=build-image /go/bin/http-service /app/http-service

# Create required directories
RUN mkdir -p /etc/http-service/certs && \
    chown -R appuser:appgroup /app /etc/http-service

# Set the working directory
WORKDIR /app

# Use the non-root user
USER appuser

# Expose the application port
EXPOSE 8000

# Set the entrypoint
ENTRYPOINT ["/app/http-service"]

# Set up a health check
HEALTHCHECK --interval=5s --timeout=3s --retries=3 \
  CMD curl -f http://localhost:8000/health || exit 1