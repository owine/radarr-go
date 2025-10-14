# Build stage
FROM golang:1.25.3-alpine@sha256:20ee0b674f987514ae3afb295b6a2a4e5fa11de8cc53a289343bbdab59b0df59 AS builder

# Build arguments for version information
ARG VERSION="dev"
ARG COMMIT="unknown"
ARG BUILD_DATE="unknown"

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with version information (pure-Go, no CGO needed)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${BUILD_DATE}'" \
    -o radarr ./cmd/radarr

# Final stage
FROM alpine:3.22.2@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412

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
