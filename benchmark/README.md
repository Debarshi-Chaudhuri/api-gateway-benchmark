# API Gateway Benchmark Tool

This directory contains a benchmarking tool and testing scenarios for comparing the performance and features of Tyk and KrakenD API gateways.

## Overview

The benchmark tool tests the following gateway features:

- **Authentication**: JWT-based authentication
- **Resilience**: Circuit breakers, timeouts, and error handling
- **Rate Limiting**: Global, endpoint, and key-based rate limiting
- **API Proxying**: HTTP and gRPC proxy capabilities

## Usage

### Running the Benchmark Tool

The benchmark tool can be run in Docker as part of the docker-compose setup:

```bash
docker-compose run benchmark
```

### Command Line Options

```
--concurrency INT     Number of concurrent clients (default: 10)
--requests INT        Total number of requests (default: 1000)
--scenario STRING     Test scenario (auth, resilience, ratelimit, proxy, all)
--grpc BOOL           Enable gRPC tests (default: true)
--output STRING       Output file for results (default: results/benchmark_results.json)
--test BOOL           Run Go tests instead of benchmarks (default: false)
```

### Running Specific Scenarios

Run only authentication tests:
```bash
docker-compose run benchmark --scenario=auth
```

Run only rate limit tests:
```bash
docker-compose run benchmark --scenario=ratelimit
```

Run only API proxying tests:
```bash
docker-compose run benchmark --scenario=proxy
```

Run only resilience tests:
```bash
docker-compose run benchmark --scenario=resilience
```

### Running Tests

To run the integration and unit tests:
```bash
docker-compose run benchmark --test=true
```

### Using the Makefile

A Makefile is provided for common operations:

```bash
# Inside the container or with Go installed locally
make test          # Run all tests
make benchmark     # Run all benchmarks
make auth          # Run auth benchmarks
make proxy         # Run proxy benchmarks
make resilience    # Run resilience benchmarks
make ratelimit     # Run rate limit benchmarks
make integration   # Run integration tests only
make clean         # Clean results directory
```

## Test Scenarios

### Authentication Tests
Tests JWT authentication for both HTTP and gRPC endpoints.

### Rate Limiting Tests
Tests the rate limiting capabilities of both gateways.

### Resilience Tests
Tests how gateways handle service failures, timeouts, and circuit breaking.

### Proxy Tests
Tests the basic API proxying capabilities for both HTTP and gRPC.

### Integration Tests
End-to-end tests that verify all components work together correctly.

## Results

Benchmark results are saved as JSON files in the `results/` directory. Example:

```json
[
  {
    "Gateway": "tyk",
    "Scenario": "auth",
    "Protocol": "HTTP",
    "RequestCount": 1000,
    "SuccessCount": 980,
    "FailureCount": 20,
    "TotalDuration": 12500000000,
    "AvgResponseTime": 125000000,
    "MinResponseTime": 50000000,
    "MaxResponseTime": 500000000,
    "RPS": 78.5
  },
  ...
]
```

The tool also produces a summary table in the console output:

```
+----------+------------+----------+----------+---------+--------+-----------+-------+
| GATEWAY  | SCENARIO   | PROTOCOL | REQUESTS | SUCCESS | FAILED | AVG TIME  | RPS   |
+----------+------------+----------+----------+---------+--------+-----------+-------+
| tyk      | auth       | HTTP     | 1000     | 980     | 20     | 125.00 ms | 78.50 |
| krakend  | auth       | HTTP     | 1000     | 990     | 10     | 110.00 ms | 89.50 |
...
+----------+------------+----------+----------+---------+--------+-----------+-------+
```