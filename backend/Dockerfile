# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files and download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install certificates and other runtime dependencies
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Create a directory for the database to ensure it's writable
RUN mkdir -p /app/data

# Set environment variables
ENV PORT=8080
ENV DB_PATH=/app/data/aggregator.db
EXPOSE 8080

# Command to run the application
CMD ["./main"]
