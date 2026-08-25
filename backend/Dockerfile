# Builder stage
FROM golang:1.26.4-alpine AS builder

WORKDIR /app

# Maximize layer caching: copy go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the worker binary statically
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/worker ./cmd/worker

# Final minimal stage
FROM alpine:3.21

# Create non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/worker .

# Use non-root user
USER appuser:appgroup

# Command to run the worker
CMD ["./worker"]
