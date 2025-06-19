# Builder stage
FROM golang:1.23-alpine AS builder

# Env variables
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Copy the rest of the code
COPY . .

# Build the binary with -w and -s flags to reduce size
RUN go build -ldflags="-w -s" -o memorydb ./cmd/main.go

# Final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/memorydb .
CMD ["./memorydb"]