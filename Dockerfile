# Build stage
FROM golang:1.23.10-alpine AS builder

# Install git for downloading dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Create appuser for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the binary with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o mcp-server ./cmd/server

# Final stage - minimal image
FROM scratch

# Import ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Import user/group files
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary
COPY --from=builder /app/mcp-server /mcp-server

# Use unprivileged user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/mcp-server", "--health-check"]

# Run the binary
ENTRYPOINT ["/mcp-server"]
