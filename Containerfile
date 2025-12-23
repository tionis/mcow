# Build Stage
FROM golang:1.24-alpine AS builder

# Install build dependencies (GCC for CGO/SQLite)
RUN apk add --no-cache build-base

WORKDIR /app

# Build the application
COPY . .
# -ldflags="-w -s" strips debug information for smaller binary
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o mc-webui .

# Runtime Stage
FROM alpine:latest

# Install runtime dependencies (CA certs for HTTPS)
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mc-webui .

# Create data directory
RUN mkdir -p data/mods

# Expose port
EXPOSE 8080

# Environment variables (Defaults, override in Quadlet/Docker)
ENV PORT=8080
ENV DB_PATH=/data/mcow.db
ENV MOD_DATA_PATH=/app/data/mods

# Mount point for data persistence
VOLUME ["/data", "/app/data/mods"]

# Run
CMD ["./mc-webui"]
