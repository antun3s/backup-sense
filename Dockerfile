# Build stage
FROM golang:1.21-alpine AS builder

# Set build arguments (will be passed from buildx)
ARG TARGETOS
ARG TARGETARCH

# Set working directory
WORKDIR /app

# Copy source files
COPY . .

# Build binary with optimizations
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /backup-sense backup-sense.go

# Runtime stage
FROM alpine:latest

# Install CA certificates for TLS support
RUN apk --no-cache add ca-certificates

# Create backup directory
RUN mkdir -p /backup

# Copy pre-built binary
COPY --from=builder /backup-sense /usr/local/bin/backup-sense

# Set environment variables
ENV BACKUP_DIR=/backup
ENV PORT=80
ENV MAX_MB=10

# Expose default port
EXPOSE 80

# Set entrypoint
ENTRYPOINT ["backup-sense"]

# Default command-line flags
CMD ["-p", "80", "-f", "/backup", "-m", "10"]