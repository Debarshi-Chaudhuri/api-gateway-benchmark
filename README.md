# API Gateway Benchmark: Tyk vs KrakenD

This project sets up a comprehensive benchmark environment to compare Tyk and KrakenD API gateways, focusing on:

- **Authentication**: JWT-based authentication
- **Resilience**: Circuit breakers, timeouts, and error handling
- **Rate Limiting**: Global, endpoint, and key-based rate limiting
- **API Proxying**: HTTP and gRPC proxy capabilities

## Project Structure

```
api-gateway-benchmark/
├── backend-services/          # Mock services to be proxied
│   ├── http-service/          # HTTP backend service
│   └── grpc-service/          # gRPC backend service
├── tyk/                      # Tyk gateway configuration
├── krakend/                  # KrakenD gateway configuration
├── benchmark/                # Benchmarking tool
└── docker-compose.yml        # Docker setup for all components
```

## Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

## Setup Instructions

1. Clone the repository:

```bash
git clone https://github.com/example/api-gateway-benchmark.git
cd api-gateway-benchmark
```

2. Build and start all services:

```bash
docker-compose build
docker-compose up -d
```

3. Verify that all services are running:

```bash
docker-compose ps
```

## Running Benchmarks

The benchmark tool is included in the Docker Compose setup. You can run it with:

```bash
docker-compose run benchmark
```

### Custom Benchmark Options

```bash
docker-compose run benchmark --help
```

Available options:
- `--concurrency`: Number of concurrent clients (default: 10)
- `--requests`: Total number of requests (default: 1000)
- `--scenario`: Test scenario (auth, resilience, ratelimit, proxy, all)
- `--grpc`: Enable gRPC tests (default: true)
- `--output`: Output file for results (default: results/benchmark_results.json)

Example:
```bash
docker-compose run benchmark --concurrency=20 --requests=5000 --scenario=auth
```

## Service Endpoints

### Tyk Gateway
- HTTP API: http://localhost:8080/http-api/
- gRPC API: localhost:8080 (grpc://localhost:8080/grpc-api/)

### KrakenD Gateway
- HTTP API: http://localhost:8081/http/
- gRPC API: http://localhost:8081/grpc/ (through HTTP-to-gRPC gateway)

## Test and Debug

### HTTP Testing

```bash
# Test Tyk HTTP endpoint
curl http://localhost:8080/http-api/

# Test KrakenD HTTP endpoint
curl http://localhost:8081/http/data
```

### gRPC Testing

A simple gRPC client is provided in `tools/grpc-client/`:

```bash
# Test Tyk gRPC endpoint
go run grpc-client.go --gateway=tyk

# Test KrakenD gRPC endpoint
go run grpc-client.go --gateway=krakend
```

## Cleanup

```bash
docker-compose down -v
```

## Notes

- The default JWT secret key for both gateways is `test-secret-key-for-benchmark`
- The benchmarking tool generates temporary JWT tokens for authentication tests
- Metrics for Tyk are available through its dashboard (not included in this setup)
- KrakenD metrics are exposed on port 8090