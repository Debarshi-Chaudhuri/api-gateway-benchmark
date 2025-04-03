#!/bin/bash

# Script to generate RSA keys and create a JWT token for testing Tyk JWT authentication

echo "=== Generating RSA keys for JWT authentication ==="

# Create directory for keys if it doesn't exist
mkdir -p jwt_keys

# Generate private key
openssl genrsa -out jwt_keys/private_key.pem 2048
echo "Private key generated: jwt_keys/private_key.pem"

# Extract public key
openssl rsa -in jwt_keys/private_key.pem -pubout -out jwt_keys/public_key.pem
echo "Public key extracted: jwt_keys/public_key.pem"

# Display the public key (this should be added to the Tyk API definition)
echo -e "\n=== Public Key (add this to jwt_rsa_public_key in API definition) ==="
cat jwt_keys/public_key.pem

# Install dependencies for JWT generation if needed
if ! command -v npm &> /dev/null; then
    echo -e "\nNode.js is required to generate JWT tokens. Please install it first."
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo -e "\njq is required for JSON formatting. Please install it first."
    exit 1
fi

# Create a temporary directory for Node.js project
mkdir -p jwt_generator
cd jwt_generator

# Initialize package.json if it doesn't exist
if [ ! -f package.json ]; then
    echo '{"name":"jwt-generator","version":"1.0.0","description":"JWT Token Generator"}' > package.json
    npm install --save jsonwebtoken
fi

# Create a JavaScript file to generate JWT token
cat > generate_jwt.js << 'EOF'
const fs = require('fs');
const jwt = require('jsonwebtoken');

// Read the private key
const privateKey = fs.readFileSync('../jwt_keys/private_key.pem');

// Current time in seconds
const now = Math.floor(Date.now() / 1000);

// Create JWT payload
const payload = {
  sub: "1234567890",  // Subject (user ID)
  name: "Test User",
  iss: "api-client",  // Issuer
  iat: now,           // Issued at
  exp: now + 3600,    // Expires in 1 hour
  scope: "user",     // Scope claim for policy mapping
  pol: "5a7a5c61-e2d9-4231-8e63-a74750e6f4b4"  // Policy field
};

// Create token with the private key and RS256 algorithm
const token = jwt.sign(payload, privateKey, { 
  algorithm: 'RS256',
  header: {
    kid: "12345"  // Key ID
  }
});

console.log(JSON.stringify({ token }));
EOF

# Generate the JWT token
echo -e "\n=== Generating JWT token ==="
node generate_jwt.js > token.json

# Display the token
echo -e "\n=== JWT Token ==="
jq -r '.token' token.json

# Create a curl command example
echo -e "\n=== Example curl command ==="
echo "curl -H \"Authorization: Bearer $(jq -r '.token' token.json)\" http://localhost:8080/http-api/api/data"

# Clean up
cd ..
echo -e "\nToken saved to jwt_generator/token.json"
echo "Keep your private key (jwt_keys/private_key.pem) secure!"