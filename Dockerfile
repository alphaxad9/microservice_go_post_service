# =========================
# Stage 1: Builder
# =========================
FROM golang:1.24.13-alpine AS builder

# Install system dependencies required for compiling Kafka (CGO)
RUN apk add --no-cache \
    gcc \
    musl-dev \
    pkgconfig \
    librdkafka-dev \
    && rm -rf /var/cache/apk/*

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the source code
COPY post_service ./post_service

# Set Environment Variables for CGO and Alpine compatibility
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Build the binaries
# Note: -tags musl tells the Kafka library to use Alpine-compatible logic
RUN go build -tags musl -ldflags="-w -s" \
    -o /app/bin/post-service \
    ./post_service/cmd/internal

RUN go build -tags musl -ldflags="-w -s" \
    -o /app/bin/outbox-publisher \
    ./post_service/cmd/outbox-publisher

RUN go build -tags musl -ldflags="-w -s" \
    -o /app/bin/kafka-consumer \
    ./post_service/cmd/kafka-consumer

# =========================
# Stage 2: Runtime
# =========================
FROM alpine:3.19

WORKDIR /app

# Install runtime libraries and clean up
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    librdkafka \
    && rm -rf /var/cache/apk/*

# Set Gin to release mode for production
ENV GIN_MODE=release

# Security: Run as a non-privileged user
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

# Copy the compiled binaries from the builder stage
COPY --from=builder --chown=appuser:appuser /app/bin/post-service .
COPY --from=builder --chown=appuser:appuser /app/bin/outbox-publisher .
COPY --from=builder --chown=appuser:appuser /app/bin/kafka-consumer .

USER appuser

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["./post-service", "health"]

# Default command (you can override this in docker-compose for the other services)
CMD ["./post-service"]