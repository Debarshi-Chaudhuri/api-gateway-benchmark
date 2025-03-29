FROM golang:1.23.7-alpine AS builder

WORKDIR /app

# Copy go.mod from root
COPY go.mod ./
# Run go mod tidy to ensure go.sum is up to date
RUN go mod tidy
RUN go mod download

# Copy the HTTP service code
COPY http-service/main.go ./

# Build the application
RUN go build -o http-service main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /app/

COPY --from=builder /app/http-service .

EXPOSE 8000

CMD ["./http-service"]