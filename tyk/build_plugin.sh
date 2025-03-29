#!/bin/bash
# build_plugin.sh - Build Tyk plugin using official Tyk plugin compiler

# Echo with timestamp and color
info() {
  echo -e "\033[0;34m[$(date '+%Y-%m-%d %H:%M:%S')]\033[0m $1"
}

# Define paths - adjust these for your environment
PROJECT_ROOT="/Users/debarshi/go/src/api-gateway-benchmark"
PLUGIN_DIR="tyk/plugins"
OUTPUT_DIR="tyk/middleware"
PLUGIN_NAME="logger"

# Create middleware directory if it doesn't exist
mkdir -p "$PROJECT_ROOT/$OUTPUT_DIR"

# Change to the plugin directory
cd "$PROJECT_ROOT/$PLUGIN_DIR"

info "Initializing Go module for Tyk plugin..."
# Step 1: Initialize Go module with an appropriate name
docker run -v "$(pwd):/plugin-source" -t --workdir /plugin-source \
  --platform=linux/amd64 --entrypoint go --rm tykio/tyk-plugin-compiler:v5.8.0 \
  tyk-plugin

info "Building Tyk plugin using official Tyk plugin compiler..."

echo "$PLUGIN_NAME.so"
# Step 2: Compile the plugin
docker run --rm -v "$(pwd):/plugin-source" --platform=linux/amd64 \
  tykio/tyk-plugin-compiler:v5.8.0 \
  "$PLUGIN_NAME.so" "$(date +%s%N)"

# Move the compiled plugin to the middleware directory
# if [ -f "$PROJECT_ROOT/$PLUGIN_DIR/$PLUGIN_NAME.so" ]; then
#   mv "$PROJECT_ROOT/$PLUGIN_DIR/$PLUGIN_NAME.so" "$PROJECT_ROOT/$OUTPUT_DIR/"
#   info "✅ Plugin built successfully and moved to $OUTPUT_DIR/$PLUGIN_NAME.so"
# else
#   info "❌ Plugin build failed"
# fi