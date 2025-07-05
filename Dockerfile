# Multi-stage build for smaller image
FROM golang:1.24.4-alpine AS builder

# Install necessary packages including build tools and SQLite3
RUN apk add --no-cache ca-certificates gcc git musl-dev sqlite-dev tzdata

# Set the working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled and cache mounts
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build -ldflags '-linkmode external -extldflags "-static"' -o main .

# Runtime stage
FROM alpine:3.22.0

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Set the working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Create sessions directory
RUN mkdir -p sessions

# Set environment variables
ENV SESSION_FILE_PATH=/root/sessions/

# Expose port (not needed for this app but good practice)
EXPOSE 8080

# Run the application
CMD ["./main"] 
