# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o welltaxpro ./src/cmd/main

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/welltaxpro .

# Make binary executable
RUN chmod +x welltaxpro

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/welltaxpro", "--config", "/app/config/prod.yaml"]
