# Multi-stage build for optimized multiplatform container image
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

# Build arguments for multiplatform support
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

# Set working directory
WORKDIR /app

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the server binary with multiplatform support
# Use TARGETOS and TARGETARCH for cross-compilation
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -installsuffix cgo -ldflags="-w -s" \
    -o grpc-server server/main.go

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for TLS connections
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN addgroup -g 1001 -S grpcuser && \
    adduser -u 1001 -S grpcuser -G grpcuser

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/grpc-server .

# Change ownership to non-root user
RUN chown grpcuser:grpcuser /app/grpc-server

# Switch to non-root user
USER grpcuser

# Expose port for both gRPC and HTTP (unified port)
EXPOSE 50051

# Health check using curl for HTTP endpoint (more reliable than nc)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:50051/health || exit 1

# Run the server
CMD ["./grpc-server"]
