# Build stage
FROM golang:1.25-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c014abd AS builder

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
FROM alpine:3.22@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1

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
