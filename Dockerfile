# Build stage
FROM golang:1.22-bullseye AS builder

WORKDIR /app

# Copy dependencies first (for better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -o wave-server ./cmd/api

# Final stage
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/wave-server .
COPY --from=builder /app/migrations ./migrations

# Create non-root user
RUN useradd -u 10001 appuser
USER 10001

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 CMD [ "wget", "-q", "-O", "-", "http://localhost:8080/health" ]

# Run the application
ENTRYPOINT ["./wave-server"]