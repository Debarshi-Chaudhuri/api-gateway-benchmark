FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install protoc
RUN apk add --no-cache protobuf protobuf-dev

# Install Go gRPC tools
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

COPY go.mod go.sum ./
RUN go mod download

# Copy source code and proto files
COPY . .

# Generate protobuf code
RUN mkdir -p proto/servicepb
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/service.proto

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o grpc-service ./server

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/grpc-service .

EXPOSE 9000

CMD ["./grpc-service"]