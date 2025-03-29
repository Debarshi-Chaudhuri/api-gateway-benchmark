#!/bin/bash
# build_plugin.sh - Build Tyk plugin with specific Go version and GCC support

# Define Go version - CHANGE THIS to match Tyk's Go version
GO_VERSION="1.20.5"  # Tyk 5.0.3 uses Go 1.20.5

# Echo with timestamp and color
info() {
  echo -e "\033[0;34m[$(date '+%Y-%m-%d %H:%M:%S')]\033[0m $1"
}

# Create middleware directory if it doesn't exist
mkdir -p tyk/middleware

info "Building Tyk plugin with Go $GO_VERSION..."

# Use Docker to ensure we have the exact Go version and GCC
docker run --rm \
  -v "$(pwd):/app" \
  -w /app \
  golang:$GO_VERSION \
  bash -c "
    # Install GCC and build tools
    apt-get update && apt-get install -y gcc libc6-dev
    
    info() { echo -e '\033[0;34m[INFO]\033[0m \$1'; }
    info 'Go version:' && go version && 
    info 'Building plugin...' &&
    cd tyk/plugins && 
    CGO_ENABLED=1 go build -buildmode=plugin -o ../middleware/logger.so logger.go
  "

# Check if build succeeded
if [ -f "tyk/middleware/logger.so" ]; then
  info "✅ Plugin built successfully at tyk/middleware/logger.so"
else
  info "❌ Plugin build failed"
  exit 1
fi