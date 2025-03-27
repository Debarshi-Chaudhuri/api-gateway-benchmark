FROM tykio/tyk-gateway:v5.0.3

# Create data directory for file-based storage
RUN mkdir -p /opt/tyk-gateway/data

# Copy Tyk configuration
COPY tyk.conf /opt/tyk-gateway/tyk.conf

# Copy API definitions
COPY api_definitions/ /opt/tyk-gateway/apps/

# Create policies directory
RUN mkdir -p /opt/tyk-gateway/policies