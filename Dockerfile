# Use Debian-based image for better SQLite support
FROM golang:1.21-bullseye AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    sqlite3 \
    libsqlite3-dev

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Use Debian slim for runtime
FROM debian:bullseye-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Create database directory and set permissions
RUN mkdir -p /app/data && chown -R 1001:root /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Create a non-root user
RUN useradd -r -u 1001 -g root appuser
USER appuser

# Ensure database directory exists and is writable
VOLUME ["/app/data"]

EXPOSE 8080

CMD ["./main"]