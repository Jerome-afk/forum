# Use Debian-based image instead of Alpine for better SQLite support
FROM golang:1.21-bullseye AS builder

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o main -ldflags="-w -s" .

# Use Debian slim for runtime
FROM debian:bullseye-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Create a non-root user
RUN useradd -r -u 1001 -g root appuser
USER appuser

EXPOSE 8080

CMD ["./main"]