FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install protoc
RUN apk add --no-cache protobuf protobuf-dev make

# Install Go gRPC tools
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Copy the root go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy proto files and generate them first
COPY backend-services/grpc-service/proto/ ./backend-services/grpc-service/proto/
COPY Makefile ./

# Generate protobuf code
RUN make proto

# Copy the rest of the application
COPY . .

# Build the gRPC service
RUN go build -o grpc-service ./backend-services/grpc-service/server

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/grpc-service .

EXPOSE 9000

CMD ["./grpc-service"]