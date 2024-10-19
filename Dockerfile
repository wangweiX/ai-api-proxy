# Build stage
FROM golang:1.23.2-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o proxy ./cmd/proxy

# Run stage
FROM alpine:latest

# Install certificates
RUN apk --no-cache add ca-certificates tzdata

# Set default timezone to Shanghai
ENV TZ=Asia/Shanghai

# Set working directory
WORKDIR /app

# Create logs directory and set permissions
RUN mkdir -p /app/logs && chmod 755 /app/logs

# Copy executable file
COPY --from=builder /app/proxy .
COPY --from=builder /app/config.yaml .

# Set environment variables, specify default config file path
ENV CONFIG_PATH=/app/config.yaml
ENV PORT=3002

# Copy entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Set timezone
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

EXPOSE $PORT

# Set entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]
