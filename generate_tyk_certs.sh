#!/bin/bash
# Script to generate server-only TLS certificates (no mTLS)

set -e

# Create directories
mkdir -p tls/ca
mkdir -p tls/redis
mkdir -p tls/tyk-gateway
mkdir -p tls/tyk-pump
mkdir -p tls/http-service

# Generate CA certificate (if not already existing)
if [ ! -f tls/ca/ca.key ]; then
  openssl genrsa -out tls/ca/ca.key 4096
  openssl req -new -x509 -key tls/ca/ca.key -out tls/ca/ca.crt -days 3650 \
    -subj "/C=US/ST=State/L=City/O=API Gateway Benchmark/OU=Security/CN=Benchmark Root CA"
fi

# Generate Redis server certificate
openssl genrsa -out tls/redis/redis.key 2048
openssl req -new -key tls/redis/redis.key -out tls/redis/redis.csr \
  -subj "/C=US/ST=State/L=City/O=API Gateway Benchmark/OU=Redis/CN=redis"
openssl x509 -req -in tls/redis/redis.csr -CA tls/ca/ca.crt -CAkey tls/ca/ca.key \
  -CAcreateserial -out tls/redis/redis.crt -days 3650
cp tls/ca/ca.crt tls/redis/ca.crt

# Generate Tyk Gateway certificate
openssl genrsa -out tls/tyk-gateway/tyk.key 2048
openssl req -new -key tls/tyk-gateway/tyk.key -out tls/tyk-gateway/tyk.csr \
  -subj "/C=US/ST=State/L=City/O=API Gateway Benchmark/OU=Tyk Gateway/CN=tyk-gateway"
openssl x509 -req -in tls/tyk-gateway/tyk.csr -CA tls/ca/ca.crt -CAkey tls/ca/ca.key \
  -CAcreateserial -out tls/tyk-gateway/tyk.crt -days 3650 -extfile <(printf "subjectAltName=DNS:tyk-gateway,DNS:localhost,IP:127.0.0.1")
cp tls/ca/ca.crt tls/tyk-gateway/ca.crt

# Generate Tyk Pump certificate
openssl genrsa -out tls/tyk-pump/pump.key 2048
openssl req -new -key tls/tyk-pump/pump.key -out tls/tyk-pump/pump.csr \
  -subj "/C=US/ST=State/L=City/O=API Gateway Benchmark/OU=Tyk Pump/CN=tyk-pump"
openssl x509 -req -in tls/tyk-pump/pump.csr -CA tls/ca/ca.crt -CAkey tls/ca/ca.key \
  -CAcreateserial -out tls/tyk-pump/pump.crt -days 3650
cp tls/ca/ca.crt tls/tyk-pump/ca.crt

# Generate HTTP service certificate
openssl genrsa -out tls/http-service/http.key 2048
openssl req -new -key tls/http-service/http.key -out tls/http-service/http.csr \
  -subj "/C=US/ST=State/L=City/O=API Gateway Benchmark/OU=HTTP Service/CN=http-service"
openssl x509 -req -in tls/http-service/http.csr -CA tls/ca/ca.crt -CAkey tls/ca/ca.key \
  -CAcreateserial -out tls/http-service/http.crt -days 3650 -extfile <(printf "subjectAltName=DNS:http-service,DNS:localhost,IP:127.0.0.1")
cp tls/ca/ca.crt tls/http-service/ca.crt

echo "All server certificates generated successfully."