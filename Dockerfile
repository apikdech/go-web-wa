# Multi-stage build for smaller image
FROM golang:1.24.4-alpine AS builder

# Install necessary packages
RUN apk add --no-cache ca-certificates git tzdata

# Set the working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

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
