#!/bin/bash
# Generate new certificates for Tyk to connect to Redis

mkdir -p tls/tyk-redis

# Generate Tyk client certificate for Redis
openssl genrsa -out tls/tyk-redis/tyk-redis.key 2048
openssl req -new -key tls/tyk-redis/tyk-redis.key -out tls/tyk-redis/tyk-redis.csr -subj "/CN=tyk-gateway"
openssl x509 -req -in tls/tyk-redis/tyk-redis.csr -CA tls/ca.crt -CAkey tls/ca.key -CAcreateserial -out tls/tyk-redis/tyk-redis.crt -days 365

# Copy CA certificate
cp tls/ca.crt tls/tyk-redis/

# Create combined PEM file
cat tls/tyk-redis/tyk-redis.key tls/tyk-redis/tyk-redis.crt > tls/tyk-redis/tyk-redis.pem

echo "Tyk-Redis certificates have been generated"