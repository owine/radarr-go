# Build stage
FROM golang:1.24-alpine AS builder

# Build arguments for version information
ARG VERSION="dev"
ARG COMMIT="unknown"
ARG BUILD_DATE="unknown"

RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with version information
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${BUILD_DATE}'" \
    -o radarr ./cmd/radarr

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