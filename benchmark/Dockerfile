FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod from root
COPY go.mod ./
# Run go mod tidy to ensure go.sum is up to date
RUN go mod tidy
RUN go mod download

# Copy benchmark code
COPY benchmark/main.go ./

# Build the application
RUN go build -o benchmark main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/benchmark .

# Create a directory for results
RUN mkdir -p /results
VOLUME /results

ENTRYPOINT ["./benchmark"]