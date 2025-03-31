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

echo "Building $PLUGIN_NAME.so"
# Step 2: Compile the plugin
docker run --rm -v "$(pwd):/plugin-source" -v "$PROJECT_ROOT/$OUTPUT_DIR:/output" --platform=linux/amd64 \
  tykio/tyk-plugin-compiler:v5.8.0 \
  "$PLUGIN_NAME.so" "$(date +%s%N)"

# Move all compiled .so files to the output directory
if ls *.so 1> /dev/null 2>&1; then
  # If .so files were created in the current directory, move them all to the output
  mv *.so "$PROJECT_ROOT/$OUTPUT_DIR/"
  info "✅ Plugin(s) built successfully and moved to $OUTPUT_DIR/"
elif [ -f "$PROJECT_ROOT/$OUTPUT_DIR/$PLUGIN_NAME.so" ]; then
  # If the file was already created in the output directory
  info "✅ Plugin built successfully in $OUTPUT_DIR/$PLUGIN_NAME.so"
else
  info "❌ Plugin build failed"
fi

cd "$PROJECT_ROOT"
