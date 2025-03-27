#!/bin/bash

echo "=== API Gateway Benchmark Troubleshooting ==="
echo "Checking container status..."
docker compose ps

echo -e "\n=== Checking HTTP Service Logs ==="
docker compose logs --tail=20 http-service

echo -e "\n=== Testing HTTP Service directly ==="
docker compose exec http-service curl -v http://localhost:8000/health
echo ""

echo -e "\n=== Testing Tyk Gateway ==="
curl -v http://localhost:8080/http-api/health
echo ""

echo -e "\n=== Testing KrakenD Gateway ==="
curl -v http://localhost:8081/health
echo ""

echo -e "\n=== Checking network connectivity ==="
docker compose exec http-service wget -q -O- http://localhost:8000/health || echo "Failed to connect to self"
docker compose exec tyk-gateway wget -q -O- http://http-service:8000/health || echo "Tyk failed to connect to HTTP service"
docker compose exec krakend wget -q -O- http://http-service:8000/health || echo "KrakenD failed to connect to HTTP service"

echo -e "\n=== All container IPs ==="
docker network inspect api-gateway-benchmark_api-network -f '{{range .Containers}}{{.Name}}: {{.IPv4Address}}
{{end}}'

echo -e "\nTroubleshooting complete. Check the above output for errors."