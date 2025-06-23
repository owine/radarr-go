# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o radarr ./cmd/radarr

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata sqlite

WORKDIR /app

# Create non-root user
RUN addgroup -g 1000 radarr && \
    adduser -D -s /bin/sh -u 1000 -G radarr radarr

# Copy binary from builder stage
COPY --from=builder /app/radarr .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/config.yaml .

# Create data directory
RUN mkdir -p /data && chown -R radarr:radarr /data /app

USER radarr

EXPOSE 7878

VOLUME ["/data", "/movies"]

CMD ["./radarr", "-data", "/data"]