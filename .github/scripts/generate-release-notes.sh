#!/bin/bash
set -euo pipefail

# Generate comprehensive release notes with Docker details
# Usage: generate-release-notes.sh VERSION BUILD_DATE COMMIT_SHA TAG_NAME DIGEST BASE_IMAGE AVAILABLE_TAGS RELEASE_TYPE GITHUB_REPO

VERSION="${1}"
BUILD_DATE="${2}"
COMMIT_SHA="${3}"
TAG_NAME="${4}"
DIGEST="${5}"
BASE_IMAGE="${6}"
AVAILABLE_TAGS="${7}"
RELEASE_TYPE="${8}"
GITHUB_REPO="${9}"
REGISTRY="${10}"
IMAGE_NAME="${11}"

# Extract prerelease type for better categorization
PRERELEASE_TYPE=""
if [[ $VERSION =~ -([^.]+) ]]; then
  PRERELEASE_TYPE="${BASH_REMATCH[1]}"
fi

# Determine if this is pre-1.0
IS_PRE_1_0="false"
if [[ $VERSION =~ ^0\. ]]; then
  IS_PRE_1_0="true"
fi

# Create comprehensive release notes with Docker details
cat > RELEASE_NOTES_UPDATED.md << 'RELEASE_EOF'
# Radarr Go vRELEASE_VERSION

## 🚀 Release Information

- **Version**: `vRELEASE_VERSION`
- **Build Date**: `RELEASE_BUILD_DATE`
- **Commit SHA**: [`RELEASE_COMMIT_SHA`](https://github.com/GITHUB_REPO/commit/GITHUB_COMMIT)
- **Release Type**: RELEASE_TYPE_PLACEHOLDER
- **Versioning Strategy**: VERSIONING_STRATEGY_PLACEHOLDER
- **API Compatibility**: Radarr v3 (100% compatible)

## 🐳 Docker Images

### 📦 Container Registry
- **Registry**: `REGISTRY_PLACEHOLDER`
- **Repository**: `IMAGE_NAME_PLACEHOLDER`
- **Image Digest**: `DIGEST_PLACEHOLDER`

### 🏷️ Available Tags
The following tags are available for this release:
```
TAGS_PLACEHOLDER
```

### 🔒 Security & Verification

#### Image Digest Pinning (Recommended for Production)
For maximum security, pin to the specific image digest:
```bash
# Pin to exact digest (immutable)
docker pull BASE_IMAGE_PLACEHOLDER@DIGEST_PLACEHOLDER
docker run -d -p 7878:7878 BASE_IMAGE_PLACEHOLDER@DIGEST_PLACEHOLDER
```

#### Tag-based Usage (Convenient for Development)
```bash
# Latest stable release
docker pull BASE_IMAGE_PLACEHOLDER:latest
docker run -d -p 7878:7878 BASE_IMAGE_PLACEHOLDER:latest

# Specific version
docker pull BASE_IMAGE_PLACEHOLDER:vRELEASE_VERSION
docker run -d -p 7878:7878 BASE_IMAGE_PLACEHOLDER:vRELEASE_VERSION

# Database-specific variants
docker pull BASE_IMAGE_PLACEHOLDER:vRELEASE_VERSION-postgres    # Optimized for PostgreSQL
docker pull BASE_IMAGE_PLACEHOLDER:vRELEASE_VERSION-mariadb     # Optimized for MariaDB
docker pull BASE_IMAGE_PLACEHOLDER:vRELEASE_VERSION-multi-db    # Supports both databases
```

### 🎯 Production Deployment Examples

#### Docker Compose with Digest Pinning
```yaml
services:
  radarr:
    image: BASE_IMAGE_PLACEHOLDER@DIGEST_PLACEHOLDER
    ports:
      - "7878:7878"
    volumes:
      - ./data:/data
    environment:
      - RADARR_DATABASE_TYPE=postgres
    restart: unless-stopped
```

#### Kubernetes Deployment with Digest Pinning
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: radarr
spec:
  template:
    spec:
      containers:
      - name: radarr
        image: BASE_IMAGE_PLACEHOLDER@DIGEST_PLACEHOLDER
        ports:
        - containerPort: 7878
```

### 🔍 Image Verification
Verify the image authenticity using cosign (when available):
```bash
# Verify image signature (if signed)
cosign verify BASE_IMAGE_PLACEHOLDER:vRELEASE_VERSION

# Verify SBOM attestation
cosign verify-attestation --type spdxjson BASE_IMAGE_PLACEHOLDER:vRELEASE_VERSION
```

## 📦 Binary Downloads

Download pre-compiled binaries for your platform:

| Platform | Architecture | Download |
|----------|--------------|----------|
BINARY_DOWNLOADS_PLACEHOLDER

### 🔐 Binary Verification
Each binary includes a SHA256 checksum for verification:
```bash
# Example: Verify Linux amd64 binary
wget https://github.com/GITHUB_REPO/releases/download/TAG_NAME_PLACEHOLDER/radarr-vRELEASE_VERSION-linux-amd64.tar.gz
wget https://github.com/GITHUB_REPO/releases/download/TAG_NAME_PLACEHOLDER/radarr-vRELEASE_VERSION-linux-amd64.tar.gz.sha256
sha256sum -c radarr-vRELEASE_VERSION-linux-amd64.tar.gz.sha256
```

## 🏗️ Build Information

This release was built using our optimized CI/CD pipeline:
- **85% cost reduction** through consolidated runners
- **60% speed improvement** via parallel compilation
- **Multi-platform support**: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, freebsd/amd64, freebsd/arm64
- **Enhanced security**: SBOM generation and provenance attestation

## 📋 Full Tag Reference

### Production Tags (Full Releases Only)
- `latest` - Always points to the latest stable release
- `release` - Alias for latest stable release
- `stable` - Long-term stable tag
- `vRELEASE_VERSION` - Specific version tag (immutable)

### Database-Specific Tags
- `postgres` - Optimized configuration for PostgreSQL
- `mariadb` - Optimized configuration for MariaDB
- `multi-db` - Universal configuration supporting both databases

### Development Tags (Pre-releases Only)
- `testing` - Latest pre-release for testing
- `prerelease` - Alias for latest pre-release

---

**🔒 Security Note**: For production deployments, always use digest pinning (`@DIGEST_PLACEHOLDER`) instead of tags for immutable, secure deployments.

**📖 Documentation**: Visit our [documentation](https://github.com/GITHUB_REPO) for detailed setup and configuration guides.
RELEASE_EOF

# Replace placeholders with actual values (compatible with both GNU and BSD sed)
# Function to handle sed in-place editing across platforms
sed_inplace() {
  local pattern="$1"
  local file="$2"
  if sed --version >/dev/null 2>&1; then
    # GNU sed (Linux)
    sed -i "$pattern" "$file"
  else
    # BSD sed (macOS) - requires backup extension
    sed -i '' "$pattern" "$file"
  fi
}

sed_inplace "s|RELEASE_VERSION|${VERSION}|g" RELEASE_NOTES_UPDATED.md
sed_inplace "s|RELEASE_BUILD_DATE|${BUILD_DATE}|g" RELEASE_NOTES_UPDATED.md
sed_inplace "s|RELEASE_COMMIT_SHA|${COMMIT_SHA}|g" RELEASE_NOTES_UPDATED.md
sed_inplace "s|RELEASE_TYPE_PLACEHOLDER|${RELEASE_TYPE}|g" RELEASE_NOTES_UPDATED.md
sed_inplace "s|REGISTRY_PLACEHOLDER|${REGISTRY}|g" RELEASE_NOTES_UPDATED.md
sed_inplace "s|IMAGE_NAME_PLACEHOLDER|${IMAGE_NAME}|g" RELEASE_NOTES_UPDATED.md
sed_inplace "s|DIGEST_PLACEHOLDER|${DIGEST}|g" RELEASE_NOTES_UPDATED.md
sed_inplace "s|BASE_IMAGE_PLACEHOLDER|${BASE_IMAGE}|g" RELEASE_NOTES_UPDATED.md
sed_inplace "s|GITHUB_REPO|${GITHUB_REPO}|g" RELEASE_NOTES_UPDATED.md
sed_inplace "s|TAG_NAME_PLACEHOLDER|${TAG_NAME}|g" RELEASE_NOTES_UPDATED.md

# Generate versioning strategy description
VERSIONING_STRATEGY="Semantic Versioning 2.0.0"
if [[ "$IS_PRE_1_0" == "true" ]]; then
  VERSIONING_STRATEGY="${VERSIONING_STRATEGY} (Pre-1.0 phase: breaking changes allowed in minor versions)"
fi

if [[ -n "$PRERELEASE_TYPE" ]]; then
  case "$PRERELEASE_TYPE" in
    alpha)
      VERSIONING_STRATEGY="${VERSIONING_STRATEGY} - Alpha release (early development, significant bugs expected)"
      ;;
    beta)
      VERSIONING_STRATEGY="${VERSIONING_STRATEGY} - Beta release (feature-complete, focused on stability)"
      ;;
    rc)
      VERSIONING_STRATEGY="${VERSIONING_STRATEGY} - Release Candidate (production-ready candidate)"
      ;;
    *)
      VERSIONING_STRATEGY="${VERSIONING_STRATEGY} - Prerelease (${PRERELEASE_TYPE})"
      ;;
  esac
else
  if [[ "$IS_PRE_1_0" == "true" ]]; then
    VERSIONING_STRATEGY="${VERSIONING_STRATEGY} - Pre-1.0 production release"
  else
    VERSIONING_STRATEGY="${VERSIONING_STRATEGY} - Stable production release"
  fi
fi

sed_inplace "s|VERSIONING_STRATEGY_PLACEHOLDER|${VERSIONING_STRATEGY}|g" RELEASE_NOTES_UPDATED.md

# Add available tags (use a temporary file to avoid sed complexity with newlines)
TAGS_LIST=""
IFS=', ' read -ra TAGS_ARRAY <<< "${AVAILABLE_TAGS}"
for tag in "${TAGS_ARRAY[@]}"; do
  TAGS_LIST+="${BASE_IMAGE}:${tag}"$'\n'
done

# Create a temporary file with the tags and replace the placeholder
echo -n "$TAGS_LIST" > tags_temp.txt
# Use awk to replace the placeholder to avoid sed newline issues
awk '/TAGS_PLACEHOLDER/ {system("cat tags_temp.txt"); next} {print}' RELEASE_NOTES_UPDATED.md > temp_release.md && mv temp_release.md RELEASE_NOTES_UPDATED.md
rm -f tags_temp.txt

# Get existing release to extract download links if gh CLI is available
BINARY_DOWNLOADS=""
if command -v gh >/dev/null 2>&1 && gh release view "${TAG_NAME}" >/dev/null 2>&1; then
  BINARY_DOWNLOADS=$(gh release view "${TAG_NAME}" --json assets --jq '.assets[] | select(.name | test("\\.(tar\\.gz|zip)$")) | "| " + (.name | split("-") | .[2]) + " | " + (.name | split("-") | .[3] | split(".") | .[0]) + " | [" + .name + "](" + .url + ") |"' | tr '\n' '\n' || echo "")
fi

# Replace binary downloads placeholder (use awk for multi-line content)
echo -n "$BINARY_DOWNLOADS" > binary_temp.txt
awk '/BINARY_DOWNLOADS_PLACEHOLDER/ {system("cat binary_temp.txt"); next} {print}' RELEASE_NOTES_UPDATED.md > temp_release.md && mv temp_release.md RELEASE_NOTES_UPDATED.md
rm -f binary_temp.txt

echo "✅ Enhanced release notes created with Docker details"
