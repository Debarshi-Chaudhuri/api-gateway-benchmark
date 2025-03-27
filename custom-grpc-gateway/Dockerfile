FROM golang:1.21-alpine AS builder

# Install make and other required tools
RUN apk add --no-cache make protobuf protobuf-dev

WORKDIR /app

# Copy the root go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Install Go gRPC tools
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Copy proto files and generate them first
COPY backend-services/grpc-service/proto/ ./backend-services/grpc-service/proto/
COPY Makefile ./

# Generate protobuf code
RUN make proto

# Copy the custom gateway code
COPY custom-grpc-gateway/ ./custom-grpc-gateway/

# Build the gateway
RUN go build -o grpc-gateway ./custom-grpc-gateway

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/grpc-gateway .

EXPOSE 8090

CMD ["./grpc-gateway"]