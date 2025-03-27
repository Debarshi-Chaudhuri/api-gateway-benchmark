FROM devopsfaith/krakend:2.5.0

# Copy KrakenD configuration
COPY krakend.json /etc/krakend/krakend.json

# Copy partials if they exist
COPY partials/ /etc/krakend/partials/

# Expose the KrakenD ports
EXPOSE 8081 8090