#!/bin/bash
# Unified TLS Certificate Setup Script
# Creates a common CA and certificates for HTTP service, Redis, and Tyk Gateway

set -e # Exit on error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Creating Unified TLS Certificate Setup ===${NC}"

# Create base directory structure
mkdir -p tls/{ca,http,redis,tyk}

# --- Step 1: Create a Common Certificate Authority (CA) ---
echo -e "${GREEN}Creating Certificate Authority...${NC}"

# Generate CA private key
openssl genrsa -out tls/ca/ca.key 4096

# Generate CA certificate
openssl req -new -x509 -sha256 -key tls/ca/ca.key -out tls/ca/ca.crt -days 3650 \
    -subj "/C=US/ST=State/L=City/O=API Gateway Benchmark/OU=Security/CN=Benchmark Root CA"

# Copy CA certificate to all service directories for convenience
cp tls/ca/ca.crt tls/http/
cp tls/ca/ca.crt tls/redis/
cp tls/ca/ca.crt tls/tyk/

echo -e "${GREEN}CA certificate created successfully${NC}"

# --- Step 2: Create HTTP Service Certificates ---
echo -e "${GREEN}Creating HTTP Service certificates...${NC}"

# Create OpenSSL config for HTTP service with proper SANs
cat > tls/http/openssl.cnf <<EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[dn]
C = US
ST = State
L = City
O = API Gateway Benchmark
OU = HTTP Service
CN = http-service

[req_ext]
subjectAltName = @alt_names

[alt_names]
DNS.1 = http-service
DNS.2 = localhost
IP.1 = 127.0.0.1
EOF

# Generate HTTP service key
openssl genrsa -out tls/http/http.key 2048

# Generate HTTP service CSR with SAN extension
openssl req -new -key tls/http/http.key -out tls/http/http.csr -config tls/http/openssl.cnf

# Generate HTTP service certificate with SAN extension
openssl x509 -req -in tls/http/http.csr -CA tls/ca/ca.crt -CAkey tls/ca/ca.key -CAcreateserial \
    -out tls/http/http.crt -days 3650 -extensions req_ext -extfile tls/http/openssl.cnf

# Create combined PEM file (useful for some applications)
cat tls/http/http.key tls/http/http.crt > tls/http/http.pem

echo -e "${GREEN}HTTP Service certificates created successfully${NC}"

# --- Step 3: Create Redis Server Certificates ---
echo -e "${GREEN}Creating Redis certificates...${NC}"

# Create OpenSSL config for Redis with proper SANs
cat > tls/redis/openssl.cnf <<EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[dn]
C = US
ST = State
L = City
O = API Gateway Benchmark
OU = Redis
CN = redis

[req_ext]
subjectAltName = @alt_names

[alt_names]
DNS.1 = redis
DNS.2 = localhost
IP.1 = 127.0.0.1
EOF

# Generate Redis key
openssl genrsa -out tls/redis/redis.key 2048

# Generate Redis CSR with SAN extension
openssl req -new -key tls/redis/redis.key -out tls/redis/redis.csr -config tls/redis/openssl.cnf

# Generate Redis certificate with SAN extension
openssl x509 -req -in tls/redis/redis.csr -CA tls/ca/ca.crt -CAkey tls/ca/ca.key -CAcreateserial \
    -out tls/redis/redis.crt -days 3650 -extensions req_ext -extfile tls/redis/openssl.cnf

# Create combined PEM file (useful for Redis)
cat tls/redis/redis.key tls/redis/redis.crt > tls/redis/redis.pem

echo -e "${GREEN}Redis certificates created successfully${NC}"

# --- Step 4: Create Tyk Gateway Certificates ---
echo -e "${GREEN}Creating Tyk Gateway certificates...${NC}"

# Create OpenSSL config for Tyk with proper SANs
cat > tls/tyk/openssl.cnf <<EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[dn]
C = US
ST = State
L = City
O = API Gateway Benchmark
OU = Tyk Gateway
CN = tyk-gateway

[req_ext]
subjectAltName = @alt_names

[alt_names]
DNS.1 = tyk-gateway
DNS.2 = localhost
IP.1 = 127.0.0.1
EOF

# Generate Tyk key
openssl genrsa -out tls/tyk/tyk.key 2048

# Generate Tyk CSR with SAN extension
openssl req -new -key tls/tyk/tyk.key -out tls/tyk/tyk.csr -config tls/tyk/openssl.cnf

# Generate Tyk certificate with SAN extension
openssl x509 -req -in tls/tyk/tyk.csr -CA tls/ca/ca.crt -CAkey tls/ca/ca.key -CAcreateserial \
    -out tls/tyk/tyk.crt -days 3650 -extensions req_ext -extfile tls/tyk/openssl.cnf

# Create combined PEM file
cat tls/tyk/tyk.key tls/tyk/tyk.crt > tls/tyk/tyk.pem

echo -e "${GREEN}Tyk Gateway certificates created successfully${NC}"

# --- Step 5: Create Tyk Pump Certificates (Symlink from Tyk for simplicity) ---
echo -e "${GREEN}Setting up Tyk Pump certificates (linked to Tyk Gateway)...${NC}"

mkdir -p tls/tyk-pump
ln -sf ../tyk/tyk.key tls/tyk-pump/tyk.key
ln -sf ../tyk/tyk.crt tls/tyk-pump/tyk.crt
ln -sf ../tyk/tyk.pem tls/tyk-pump/tyk.pem
ln -sf ../ca/ca.crt tls/tyk-pump/ca.crt

echo -e "${GREEN}Tyk Pump certificate links created successfully${NC}"

# --- Step 6: Verify certificates ---
echo -e "${BLUE}Verifying certificate details:${NC}"

echo -e "\n${BLUE}CA Certificate:${NC}"
openssl x509 -noout -text -in tls/ca/ca.crt | grep "Subject:" -A 1
openssl x509 -noout -text -in tls/ca/ca.crt | grep "Issuer:" -A 1

echo -e "\n${BLUE}HTTP Service Certificate:${NC}"
openssl x509 -noout -text -in tls/http/http.crt | grep "Subject:" -A 1
openssl x509 -noout -text -in tls/http/http.crt | grep "X509v3 Subject Alternative Name:" -A 1

echo -e "\n${BLUE}Redis Certificate:${NC}"
openssl x509 -noout -text -in tls/redis/redis.crt | grep "Subject:" -A 1
openssl x509 -noout -text -in tls/redis/redis.crt | grep "X509v3 Subject Alternative Name:" -A 1

echo -e "\n${BLUE}Tyk Gateway Certificate:${NC}"
openssl x509 -noout -text -in tls/tyk/tyk.crt | grep "Subject:" -A 1
openssl x509 -noout -text -in tls/tyk/tyk.crt | grep "X509v3 Subject Alternative Name:" -A 1

# --- Step 7: Set permissions ---
echo -e "\n${GREEN}Setting proper file permissions...${NC}"
chmod 600 tls/ca/ca.key
chmod 644 tls/ca/ca.crt
chmod 600 tls/http/http.key
chmod 644 tls/http/http.crt
chmod 600 tls/redis/redis.key
chmod 644 tls/redis/redis.crt
chmod 600 tls/tyk/tyk.key
chmod 644 tls/tyk/tyk.crt

echo -e "\n${GREEN}All certificates have been created successfully!${NC}"
echo -e "${BLUE}Certificate directory structure:${NC}"
find tls -type f | sort

echo -e "\n${BLUE}Next steps:${NC}"
echo -e "1. Update service configurations to use these certificates"
echo -e "2. Update Docker Compose volume mounts to include these certificates"
echo -e "3. Restart all services"