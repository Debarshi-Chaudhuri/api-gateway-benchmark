FROM alpine AS build
COPY tyk/middleware/ /opt/tyk-gateway/middleware/

RUN chmod +x /opt/tyk-gateway/middleware/

FROM --platform=linux/amd64 tykio/tyk-gateway:v5.8.0

# Copy configuration and API definitions
COPY tyk/tyk.conf /opt/tyk-gateway/tyk.conf
COPY tyk/api_definitions/ /opt/tyk-gateway/apps/
COPY --from=build /opt/tyk-gateway/middleware/ /opt/tyk-gateway/middleware/

HEALTHCHECK --interval=5s --timeout=3s --retries=3 \
  CMD curl -f http://localhost:8080/hello || exit 1
