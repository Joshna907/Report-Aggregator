# Build stage
FROM golang:1.22-alpine AS builder

# Install build-base for SQLite (CGO)
RUN apk add --no-cache build-base

WORKDIR /app

# Copy dependency files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 go build -o report-aggregator ./cmd/server/main.go

# Final stage
FROM alpine:latest

# Install sqlite-libs for runtime
RUN apk add --no-cache sqlite-libs ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/report-aggregator .

# Ensure storage directory exists for SQLite
RUN mkdir -p /app/data

# Default environment variables
ENV GIN_MODE=release
ENV PORT=8080

EXPOSE 8080

# Command to run the application
CMD ["./report-aggregator"]
