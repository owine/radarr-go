# Docker Image Cleanup Strategy

## Current Docker Tag Issues

Based on your description, there are inconsistent Docker tags that need cleanup:
- `mariadb`, `postgres`, `multi-db` - unclear versioning
- `stable` - may not reflect actual stable status
- `v0.0.10-*` - deprecated experimental tags
- Various inconsistent tags

## Recommended Docker Tag Cleanup

### Phase 1: Identify Current Tags
```bash
# List all current tags (GitHub Container Registry)
gh api "repos/OWNER/REPO/packages/container/radarr-go/versions" \
  --jq '.[] | {id: .id, tags: .metadata.container.tags}'

# Or check Docker Hub if published there
curl -s "https://registry.hub.docker.com/v2/repositories/OWNER/radarr-go/tags/" | \
  jq '.results[] | {name: .name, last_updated: .last_updated}'
```

### Phase 2: Deprecate Old Tags

#### Mark v0.0.x Series as Deprecated
```bash
# Update package descriptions to mark as deprecated
gh api --method PATCH "repos/OWNER/REPO/packages/container/radarr-go" \
  --field description="⚠️ DEPRECATED: Use v0.9.0-alpha+ tags. v0.0.x series was experimental."
```

#### Retag Confusing Tags
For tags like `mariadb`, `postgres`, `stable` that don't have version context:

```bash
# Option 1: Delete confusing tags (recommended if safe)
# These would be deleted from the registry - be careful!

# Option 2: Retag with version context (safer)
docker pull ghcr.io/OWNER/radarr-go:postgres
docker tag ghcr.io/OWNER/radarr-go:postgres ghcr.io/OWNER/radarr-go:v0.0.10-postgres-deprecated
docker push ghcr.io/OWNER/radarr-go:v0.0.10-postgres-deprecated
```

### Phase 3: Implement Clean Tagging Strategy

#### New Tagging Rules (matches VERSIONING.md):

**Current (Pre-release) Tags:**
```bash
# Version-specific (recommended for users)
ghcr.io/OWNER/radarr-go:v0.9.0-alpha

# Database-specific variants
ghcr.io/OWNER/radarr-go:v0.9.0-alpha-postgres
ghcr.io/OWNER/radarr-go:v0.9.0-alpha-mariadb
ghcr.io/OWNER/radarr-go:v0.9.0-alpha-multi-db

# Testing tags (auto-updating)
ghcr.io/OWNER/radarr-go:testing
ghcr.io/OWNER/radarr-go:prerelease
ghcr.io/OWNER/radarr-go:alpha
```

**Future (Production) Tags:**
```bash
# Production releases (post v1.0.0)
ghcr.io/OWNER/radarr-go:latest
ghcr.io/OWNER/radarr-go:stable
ghcr.io/OWNER/radarr-go:v1.0.0

# Database-specific production
ghcr.io/OWNER/radarr-go:postgres  # latest stable with postgres
ghcr.io/OWNER/radarr-go:mariadb   # latest stable with mariadb
```

### Phase 4: CI/CD Integration

Update your release workflow to implement the new tagging strategy:

```yaml
# .github/workflows/release.yml additions
- name: Generate Docker tags
  id: meta
  uses: docker/metadata-action@v5
  with:
    images: ghcr.io/${{ github.repository }}
    tags: |
      # Version-specific tag
      type=semver,pattern=v{{version}}

      # Pre-release tags
      type=raw,value=testing,enable={{is_default_branch}}
      type=raw,value=prerelease,enable=${{ github.event.release.prerelease }}
      type=raw,value=alpha,enable=${{ contains(github.ref_name, 'alpha') }}
      type=raw,value=beta,enable=${{ contains(github.ref_name, 'beta') }}

      # Production tags (only for non-prerelease)
      type=raw,value=latest,enable=${{ !github.event.release.prerelease }}
      type=raw,value=stable,enable=${{ !github.event.release.prerelease }}

- name: Build and push Docker images
  uses: docker/build-push-action@v5
  with:
    platforms: linux/amd64,linux/arm64
    push: true
    tags: ${{ steps.meta.outputs.tags }}
    labels: ${{ steps.meta.outputs.labels }}

    # Build database-specific variants
    targets: |
      base
      postgres
      mariadb
      multi-db
```

### Phase 5: Database-Specific Images

Create optimized images for different database backends:

```dockerfile
# Dockerfile.multi-stage
FROM golang:1.23-alpine AS builder
# ... build process ...

# Base image
FROM alpine:3.19 AS base
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/radarr /usr/local/bin/
ENTRYPOINT ["radarr"]

# PostgreSQL optimized
FROM base AS postgres
LABEL org.opencontainers.image.description="Radarr Go - PostgreSQL optimized"
ENV RADARR_DATABASE_TYPE=postgres
ENV RADARR_DATABASE_PORT=5432

# MariaDB optimized
FROM base AS mariadb
LABEL org.opencontainers.image.description="Radarr Go - MariaDB optimized"
ENV RADARR_DATABASE_TYPE=mariadb
ENV RADARR_DATABASE_PORT=3306

# Multi-database support
FROM base AS multi-db
LABEL org.opencontainers.image.description="Radarr Go - Multi-database support"
# No default database type - user configures
```

## Safe Cleanup Commands

### For v0.9.0-alpha Release

```bash
# Build new properly tagged images
docker build -t ghcr.io/OWNER/radarr-go:v0.9.0-alpha .
docker build -t ghcr.io/OWNER/radarr-go:v0.9.0-alpha-postgres --target postgres .
docker build -t ghcr.io/OWNER/radarr-go:v0.9.0-alpha-mariadb --target mariadb .

# Tag testing versions
docker tag ghcr.io/OWNER/radarr-go:v0.9.0-alpha ghcr.io/OWNER/radarr-go:testing
docker tag ghcr.io/OWNER/radarr-go:v0.9.0-alpha ghcr.io/OWNER/radarr-go:alpha

# Push new tags
docker push ghcr.io/OWNER/radarr-go:v0.9.0-alpha
docker push ghcr.io/OWNER/radarr-go:v0.9.0-alpha-postgres
docker push ghcr.io/OWNER/radarr-go:v0.9.0-alpha-mariadb
docker push ghcr.io/OWNER/radarr-go:testing
docker push ghcr.io/OWNER/radarr-go:alpha
```

### User Communication Strategy

#### Docker Hub/GitHub Packages Description Updates:
```markdown
# Radarr Go - Movie Collection Manager

Complete rewrite of Radarr in Go with 100% API compatibility.

## Current Status: v0.9.0-alpha (95% feature parity)

### Recommended Tags:
- `v0.9.0-alpha`: Specific alpha version (recommended)
- `testing`: Latest pre-release (auto-updating)
- `alpha`: Latest alpha release (auto-updating)

### Database Variants:
- `v0.9.0-alpha-postgres`: PostgreSQL optimized (recommended)
- `v0.9.0-alpha-mariadb`: MariaDB optimized
- `v0.9.0-alpha-multi-db`: Supports both databases

### ⚠️ Deprecated Tags:
- `v0.0.10`, `v0.0.10-beta.1`: Experimental series (deprecated)
- `mariadb`, `postgres`, `stable`: Use versioned tags instead

### Migration:
Users on v0.0.x should migrate to v0.9.0-alpha (fresh install required).
See MIGRATION.md for details.
```

## User Migration Commands

### For Current v0.0.x Users:
```bash
# Stop old version
docker-compose down

# Backup data (optional)
docker run --rm -v radarr_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/radarr-data-backup.tar.gz /data

# Update docker-compose.yml
# Change image from old tag to: ghcr.io/OWNER/radarr-go:v0.9.0-alpha

# Start new version (fresh installation)
docker-compose up -d
```

### For Users with Confusing Tags:
```bash
# If currently using unversioned tags like 'postgres', 'stable', etc.
# Update to properly versioned tags:

# From: ghcr.io/OWNER/radarr-go:postgres
# To:   ghcr.io/OWNER/radarr-go:v0.9.0-alpha-postgres

docker pull ghcr.io/OWNER/radarr-go:v0.9.0-alpha-postgres
# Update docker-compose.yml with new tag
docker-compose up -d
```

## Registry Cleanup Checklist

### Safe Actions (Recommended):
- ✅ Update package descriptions with deprecation notices
- ✅ Create new properly versioned tags
- ✅ Update CI/CD to use new tagging strategy
- ✅ Communicate changes to users before cleanup

### Potentially Breaking Actions (Careful!):
- ⚠️ Delete old unversioned tags (`mariadb`, `postgres`, `stable`)
- ⚠️ Delete experimental `v0.0.x` tags
- ⚠️ Only do after confirming no active users

### Verification Steps:
```bash
# Verify new tags exist and work
docker pull ghcr.io/OWNER/radarr-go:v0.9.0-alpha
docker run --rm ghcr.io/OWNER/radarr-go:v0.9.0-alpha --version

# Check tag metadata
docker inspect ghcr.io/OWNER/radarr-go:v0.9.0-alpha | jq '.[0].Config.Labels'

# Verify database variants
docker run --rm ghcr.io/OWNER/radarr-go:v0.9.0-alpha-postgres env | grep RADARR_DATABASE
```

## Long-term Tag Management

### Automated Cleanup (Future):
```bash
# Script to clean up old pre-release tags (keep last 5)
#!/bin/bash
gh api "repos/OWNER/REPO/packages/container/radarr-go/versions" \
  --jq '.[] | select(.metadata.container.tags[] | contains("alpha")) | .id' | \
  tail -n +6 | \
  xargs -I {} gh api --method DELETE "repos/OWNER/REPO/packages/container/radarr-go/versions/{}"
```

### Retention Policy:
- **Production releases**: Keep all (v1.0.0+)
- **Pre-releases**: Keep last 10 versions
- **Testing tags**: Always current (auto-updating)
- **Deprecated**: Mark clearly, remove after 6 months

---

**Implementation Priority**:
1. Create v0.9.0-alpha tags immediately
2. Update CI/CD for future releases
3. Communicate changes to users
4. Clean up confusing tags after grace period
5. Implement automated cleanup for pre-releases
