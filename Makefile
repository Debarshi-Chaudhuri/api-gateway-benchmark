.PHONY: proto build docker-build

all: proto build

# Generate protobuf code
proto:
	mkdir -p backend-services/grpc-service/proto/servicepb
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		backend-services/grpc-service/proto/service.proto

# Build all services
build:
	go build -o bin/grpc-service ./backend-services/grpc-service/server
	go build -o bin/http-service ./backend-services/http-service
	go build -o bin/benchmark ./benchmark
	go build -o bin/grpc-client ./tools/grpc-client

# Build Docker images
docker-build:
	docker-compose build