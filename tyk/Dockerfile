FROM golang:1.20.5-bullseye AS builder 

# Install necessary build tools
RUN apt-get update && apt-get install -y \
    gcc g++ make musl-dev libc-dev linux-libc-dev binutils build-essential

WORKDIR /go/src/api-gateway-benchmark

# Copy dependencies
COPY go.mod ./
COPY tyk/plugins ./tyk/plugins

# Create middleware directory
RUN mkdir -p ./tyk/middleware

# Ensure correct compiler paths
RUN cd ./tyk/plugins && \
    CGO_ENABLED=1 CC="gcc" CXX="g++" go build -v -x -buildmode=plugin -o ../middleware/logger.so logger.go

FROM tykio/tyk-gateway:v5.0.3

# Copy configuration and API definitions
COPY tyk/tyk.conf /opt/tyk-gateway/tyk.conf
RUN mkdir -p /opt/tyk-gateway/apps /opt/tyk-gateway/middleware
COPY tyk/api_definitions/ /opt/tyk-gateway/apps/
COPY --from=builder /go/src/api-gateway-benchmark/tyk/middleware/ /opt/tyk-gateway/middleware/
