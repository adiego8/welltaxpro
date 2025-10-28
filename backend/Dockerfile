# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o welltaxpro ./src/cmd/server

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/welltaxpro .

# Copy production config template (env vars will be substituted at runtime)
COPY config/environment/prod-config.yaml /app/config/prod.yaml

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/welltaxpro", "--config", "/app/config/prod.yaml"]
