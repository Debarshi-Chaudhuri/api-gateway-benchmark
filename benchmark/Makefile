.PHONY: all test benchmark auth proxy resilience ratelimit integration clean

all: test benchmark

# Run all tests
test:
	go test -v ./scenarios/...

# Run all benchmarks
benchmark:
	go run main.go --concurrency=10 --requests=1000 --output=results/benchmark_results.json

# Run specific benchmark tests
auth:
	go run main.go --scenario=auth --concurrency=10 --requests=500 --output=results/auth_results.json

proxy:
	go run main.go --scenario=proxy --concurrency=10 --requests=500 --output=results/proxy_results.json

resilience:
	go run main.go --scenario=resilience --concurrency=5 --requests=100 --output=results/resilience_results.json

ratelimit:
	go run main.go --scenario=ratelimit --concurrency=20 --requests=1000 --output=results/ratelimit_results.json

# Run integration tests only
integration:
	go test -v ./scenarios/integration_test.go

# Clean results
clean:
	rm -rf results/*.json