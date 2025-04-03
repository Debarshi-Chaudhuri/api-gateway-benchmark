#!/bin/bash
# Script to generate improved TLS certificates for HTTP Service

# Create directories if they don't exist
mkdir -p tls/http

# Generate CA certificate if it doesn't exist
if [ ! -f tls/ca.key ]; then
  echo "Generating CA certificate..."
  openssl genrsa -out tls/ca.key 4096
  openssl req -new -x509 -key tls/ca.key -out tls/ca.crt -days 365 -subj "/CN=API Gateway Benchmark CA"
fi

# Generate HTTP service certificate with proper Subject Alternative Names
echo "Generating HTTP service certificate with proper SANs..."

# Create OpenSSL config file for SAN
cat > tls/http/openssl.cnf <<EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[dn]
CN = http-service

[req_ext]
subjectAltName = @alt_names

[alt_names]
DNS.1 = http-service
DNS.2 = localhost
IP.1 = 127.0.0.1
EOF

# Generate HTTP service key and CSR with SAN extension
openssl genrsa -out tls/http/http.key 2048
openssl req -new -key tls/http/http.key -out tls/http/http.csr -config tls/http/openssl.cnf

# Generate HTTP service certificate with SAN extension
openssl x509 -req -in tls/http/http.csr -CA tls/ca.crt -CAkey tls/ca.key -CAcreateserial \
  -out tls/http/http.crt -days 365 -extensions req_ext -extfile tls/http/openssl.cnf

# Create combined PEM file
cat tls/http/http.key tls/http/http.crt > tls/http/http.pem

# Copy CA certificate
cp tls/ca.crt tls/http/

echo "HTTP service certificates have been generated in the tls/http directory"
echo "These certificates include proper Subject Alternative Names for validation"