# Build stage
FROM golang:1.23.2-bookworm AS builder

# Enable CGO
ENV CGO_ENABLED=1

# Set working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy the source code
COPY . .

# Build the application
RUN go build -o server main.go app.go

# Runtime stage
FROM debian:bookworm-slim

# Install required runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates sqlite3 && \
    rm -rf /var/lib/apt/lists/*

# Set working directory inside the container
WORKDIR /app

# Copy the application binary from the builder stage
COPY --from=builder /app/server .

# Copy the config file
COPY config.yaml .
# Set Gin to production mode
ENV GIN_MODE=release


# Set the database file location as an environment variable (optional)
ENV DATABASE_FILE=fingerprints.db

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./server"]