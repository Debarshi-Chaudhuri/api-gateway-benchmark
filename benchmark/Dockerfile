FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o benchmark .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/benchmark .

# Create a directory for results
RUN mkdir -p /results
VOLUME /results

ENTRYPOINT ["./benchmark"]