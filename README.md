# Simple API Gateway Benchmark: Tyk vs KrakenD

This project sets up a simple environment to benchmark Tyk and KrakenD API gateways for HTTP requests. It uses a single backend service and tests HTTP request throughput and response times for each gateway.

## Project Structure

```
api-gateway-benchmark/
├── go.mod                   # Single Go module file
├── http-service/            # HTTP backend service
│   ├── Dockerfile
│   └── main.go
├── benchmark/               # Benchmarking tool
│   ├── Dockerfile
│   └── main.go
├── tyk/                     # Tyk gateway configuration
│   ├── Dockerfile
│   ├── tyk.conf
│   └── api_definitions/
│       └── http_api.json
├── krakend/                 # KrakenD gateway configuration
│   ├── Dockerfile
│   └── krakend.json
├── docker-compose.yml       # Docker setup for all components
└── README.md                # Project documentation
```

## Architecture

- **Backend Service**: A simple HTTP service that responds to requests
- **API Gateways**: 
  - Tyk Gateway
  - KrakenD Gateway
- **Benchmark Tool**: A tool to send concurrent HTTP requests and measure performance

## Prerequisites

- Docker and Docker Compose

## Setup Instructions

1. Clone the repository:

```bash
git clone https://github.com/example/simple-api-gateway-benchmark.git
cd simple-api-gateway-benchmark
```

2. Build and start all services:

```bash
docker-compose build
docker-compose up -d
```

3. Wait for all services to start (this might take a few seconds)

4. Run the benchmark:

```bash
docker-compose run benchmark
```

## Benchmark Options

```bash
docker-compose run benchmark --help
```

Available options:
- `--concurrency`: Number of concurrent clients (default: 10)
- `--requests`: Total number of requests (default: 1000)
- `--timeout`: Request timeout (default: 30s)
- `--output`: Output file for results (default: results/benchmark_results.txt)

Example:
```bash
docker-compose run benchmark --concurrency=20 --requests=5000
```

## Service Endpoints

### Backend Service (not directly accessible through docker-compose network)
- Root: http://http-service:8000/
- Data: http://http-service:8000/api/data
- Health: http://http-service:8000/health

### Tyk Gateway
- API: http://localhost:8080/http-api/api/data

### KrakenD Gateway
- API: http://localhost:8081/http/data

## Testing Individual Services

You can test each gateway directly using curl:

```bash
# Test Tyk Gateway
curl http://localhost:8080/http-api/api/data

# Test KrakenD Gateway
curl http://localhost:8081/http/data
```

## Interpretation of Results

The benchmark tool provides the following metrics:
- Total requests
- Successful requests
- Failed requests
- Average response time
- Minimum response time
- Maximum response time 
- Requests per second (RPS)

Higher RPS values and lower response times generally indicate better performance.

## Cleanup

When you're done with benchmarking, you can stop and remove all containers with:

```bash
docker-compose down
```