FROM golang:1.23.7-alpine

WORKDIR /app

# Copy everything needed for the build
COPY go.mod go.sum ./
COPY benchmark/main.go ./main.go

# Download dependencies and build
RUN go mod download
RUN go build -o benchmark main.go

# Create results directory
RUN mkdir -p /results
VOLUME /results

ENTRYPOINT ["./benchmark"]