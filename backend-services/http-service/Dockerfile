FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy the root go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN go build -o http-service ./backend-services/http-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

COPY --from=builder /app/http-service .

EXPOSE 8000

CMD ["./http-service"]