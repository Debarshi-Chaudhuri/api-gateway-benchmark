FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy the root go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN go build -o benchmark ./benchmark

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/benchmark .

# Create a directory for results
RUN mkdir -p /results
VOLUME /results

ENTRYPOINT ["./benchmark"]