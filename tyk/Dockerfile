FROM tykio/tyk-gateway:v5.0.3

# Copy Tyk configuration
COPY tyk.conf /opt/tyk-gateway/tyk.conf

# Create apps directory
RUN mkdir -p /opt/tyk-gateway/apps

# Copy API definitions
COPY api_definitions/ /opt/tyk-gateway/apps/