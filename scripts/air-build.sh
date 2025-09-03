#!/bin/bash

# Air build script for development
# This script handles the build process for Air hot reload

set -e

# Get version info
VERSION="dev"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Build the binary
CGO_ENABLED=0 go build \
    -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}" \
    -gcflags="-N -l" \
    -o ./tmp/radarr-dev \
    ./cmd/radarr

echo "Build completed successfully"
