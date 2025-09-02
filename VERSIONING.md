# Versioning Strategy for Radarr Go

## Overview

This document defines the comprehensive versioning strategy for Radarr Go, a complete rewrite of the Radarr movie collection manager from C#/.NET to Go. Our versioning approach ensures compatibility, predictability, and clear communication of changes to users while supporting our goal of 100% API compatibility with Radarr v3.

## Current Status

- **Current Version**: v0.9.0-alpha
- **Feature Parity**: 95% complete with original Radarr
- **Target**: 100% API compatibility with Radarr v3
- **Pre-1.0 Status**: Currently in alpha/beta phase approaching production readiness

## Semantic Versioning (SemVer) Framework

Radarr Go follows [Semantic Versioning 2.0.0](https://semver.org/) with project-specific adaptations:

### Version Format
```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

### Component Definitions

#### MAJOR (X.y.z)
**Pre-1.0 Strategy** (Current Phase):
- **0.y.z**: Breaking changes are allowed in minor versions
- Major version remains 0 until API stability is achieved
- Breaking changes must be documented in CHANGELOG.md

**Post-1.0 Strategy** (Future):
- Incremented for incompatible API changes
- Breaking changes to public Go APIs
- Removal of deprecated features
- Changes that break API compatibility with Radarr v3

#### MINOR (x.Y.z)
**Pre-1.0 Strategy**:
- New functionality with potential breaking changes
- Major feature additions
- Significant architectural improvements
- Database schema changes

**Post-1.0 Strategy**:
- Backward-compatible new functionality
- New API endpoints or parameters
- Feature enhancements
- Performance improvements

#### PATCH (x.y.Z)
**Both Pre/Post-1.0**:
- Backward-compatible bug fixes
- Security patches
- Performance optimizations without API changes
- Documentation updates
- Internal refactoring

### Pre-release Identifiers

#### Alpha (-alpha, -alpha.N)
```bash
# Examples
v0.9.0-alpha
v0.10.0-alpha.1
v1.0.0-alpha.2
```

**Characteristics**:
- Early development versions
- May have incomplete features
- Significant bugs expected
- Breaking changes likely
- Not suitable for production

**Release Criteria**:
- Core functionality implemented
- Basic testing completed
- Major architectural decisions made

#### Beta (-beta, -beta.N)
```bash
# Examples
v0.9.0-beta.1
v1.0.0-beta.3
```

**Characteristics**:
- Feature-complete for the target release
- Focused on stability and bug fixes
- API changes minimized
- Suitable for testing environments

**Release Criteria**:
- All planned features implemented
- Comprehensive testing completed
- Performance benchmarks met
- Documentation updated

#### Release Candidate (-rc, -rc.N)
```bash
# Examples
v1.0.0-rc.1
v1.2.0-rc.2
```

**Characteristics**:
- Production-ready candidates
- Only critical bug fixes
- No new features
- API frozen

**Release Criteria**:
- All tests passing
- Performance requirements met
- Security audit completed
- Documentation finalized

## Docker Image Versioning Strategy

### Tag Hierarchy

#### Version-Specific Tags
```bash
# Exact version
ghcr.io/radarr/radarr-go:v1.2.3

# Database-specific tags
ghcr.io/radarr/radarr-go:v1.2.3-postgres
ghcr.io/radarr/radarr-go:v1.2.3-mariadb
ghcr.io/radarr/radarr-go:v1.2.3-multi-db
```

#### Release Type Tags

**Production Releases** (non-prerelease):
```bash
# Latest stable
ghcr.io/radarr/radarr-go:latest
ghcr.io/radarr/radarr-go:release
ghcr.io/radarr/radarr-go:stable

# Stable with version
ghcr.io/radarr/radarr-go:stable-v1.2.3

# Calendar versioning
ghcr.io/radarr/radarr-go:2025.01

# Database focus
ghcr.io/radarr/radarr-go:postgres
ghcr.io/radarr/radarr-go:mariadb
ghcr.io/radarr/radarr-go:multi-db
```

**Pre-release** (alpha/beta/rc):
```bash
# Testing versions
ghcr.io/radarr/radarr-go:testing
ghcr.io/radarr/radarr-go:testing-v1.2.3-beta.1
ghcr.io/radarr/radarr-go:prerelease
```

### Tag Assignment Rules

| Release Type | Tags Applied | Production Ready | Auto-Update Recommended |
|-------------|-------------|------------------|------------------------|
| Stable Release | `latest`, `release`, `stable`, `v1.2.3`, database variants | ✅ Yes | ✅ Yes |
| Pre-release | `testing`, `prerelease`, `v1.2.3-beta.1` | ❌ No | ❌ No |

## Version Bump Criteria

### Major Version Bumps (Breaking Changes)

**Pre-1.0**: Reserved for 1.0.0 release
- Achievement of 100% API parity with Radarr v3
- Production readiness milestone
- API stability guarantee

**Post-1.0**: Incompatible changes
- Breaking API changes
- Removal of deprecated features
- Major architectural redesign
- Database schema incompatibilities

### Minor Version Bumps (Features)

**Pre-1.0**: New functionality
- Significant feature additions
- API endpoint additions
- Major performance improvements
- Database schema changes

**Post-1.0**: Compatible enhancements
- New features maintaining backward compatibility
- API extensions
- Performance improvements
- New integrations

### Patch Version Bumps (Fixes)

**All Versions**: Bug fixes and minor improvements
- Security patches
- Bug fixes
- Performance optimizations
- Documentation updates
- Internal refactoring

## Automation and Validation

### GitHub Actions Integration

Our CI/CD pipeline enforces versioning standards:

```yaml
# Version validation in release workflow
- name: Validate version format
  run: |
    if [[ ! $VERSION =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)(-[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*)?$ ]]; then
      echo "❌ Invalid version format: $VERSION"
      echo "Expected: X.Y.Z or X.Y.Z-prerelease"
      exit 1
    fi
```

### Build Information Injection

Version information is embedded at build time:

```go
// Build information - set by ldflags
var (
    version = "dev"      // Set to actual version during build
    commit  = "unknown"  // Git commit SHA
    date    = "unknown"  // Build timestamp
)
```

Build command:
```bash
LDFLAGS="-w -s -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${BUILD_DATE}'"
go build -ldflags="${LDFLAGS}" ./cmd/radarr
```

## Release Process

### 1. Development Phase
```bash
# Feature development on feature branches
git checkout -b feature/new-functionality
# ... development work ...
git commit -m "feat(api): add new movie search endpoint"
```

### 2. Version Planning
```bash
# Determine version bump type based on changes
# Check CHANGELOG.md for accumulated changes
# Decide: patch (bug fixes), minor (features), major (breaking)
```

### 3. Pre-release Process
```bash
# Create pre-release for testing
git tag v0.9.1-beta.1
git push origin v0.9.1-beta.1

# GitHub Actions will:
# - Validate version format
# - Build multi-platform binaries
# - Create Docker images with testing tags
# - Run integration tests
```

### 4. Release Preparation
```bash
# Update CHANGELOG.md with release notes
# Update version references in documentation
# Final testing and validation
```

### 5. Production Release
```bash
# Create production release
git tag v0.9.1
git push origin v0.9.1

# Or use GitHub UI to create release
# GitHub Actions will:
# - Build and publish release artifacts
# - Create Docker images with production tags
# - Run comprehensive integration tests
# - Update release notes with Docker details
```

### 6. Post-Release
```bash
# Verify release artifacts
# Monitor for issues
# Update documentation if needed
```

## API Compatibility Promise

### Radarr v3 API Compatibility

**Current Goal**: 100% compatibility with Radarr v3 API
- **Status**: 95% complete (v0.9.0-alpha)
- **Commitment**: No breaking changes to Radarr v3 API endpoints
- **Testing**: Automated compatibility testing in CI/CD

**Compatibility Matrix**:

| Component | Status | Notes |
|----------|--------|-------|
| Core API Endpoints | ✅ Complete | 150+ endpoints implemented |
| Authentication | ✅ Complete | API key compatibility |
| Database Models | ✅ Complete | GORM-based with validation |
| WebSocket Events | 🔄 In Progress | Real-time updates |
| Legacy Endpoints | ✅ Complete | Backward compatibility maintained |

### Go API Versioning

**Pre-1.0**: Go API changes allowed
- Internal APIs may change
- Public package APIs evolving
- Import paths stable

**Post-1.0**: Go module compatibility
- Semantic import versioning
- Backward compatibility for public APIs
- Clear deprecation process

## Version History and Migration

### Notable Releases

| Version | Type | Key Changes | Migration Notes |
|---------|------|------------|----------------|
| v0.0.10 | Alpha | Initial public release | Database initialization required |
| v0.0.10-beta.1 | Beta | Enhanced Docker support | Environment variable changes |
| v0.9.0-alpha | Alpha | Near-feature parity | Configuration format updates |

### Upcoming Milestones

| Target Version | Goals | Timeline |
|---------------|-------|----------|
| v0.9.1 | Bug fixes, stability | Q1 2025 |
| v0.10.0 | 100% feature parity | Q1 2025 |
| v1.0.0 | Production release | Q2 2025 |

## Development Workflow Integration

### Branch Naming
```bash
# Feature branches
feature/add-notification-provider
feature/improve-search-performance

# Bug fix branches
fix/database-connection-leak
fix/api-authentication-issue

# Release branches
release/v0.9.1
release/v1.0.0
```

### Commit Message Format
```bash
# Type(scope): description
feat(api): add movie collection management endpoints
fix(database): resolve connection pool exhaustion
docs(versioning): update release process documentation
chore(deps): update Go dependencies
```

### Tag Creation
```bash
# Lightweight tags for development
git tag v0.9.1-alpha.1

# Annotated tags for releases
git tag -a v0.9.1 -m "Release v0.9.1: Stability improvements and bug fixes"
```

## Monitoring and Metrics

### Version Adoption Tracking
- Docker image pull statistics by tag
- GitHub release download counts
- API usage metrics by version

### Performance Benchmarks
- Automated benchmark testing in CI/CD
- Performance regression detection
- Resource utilization tracking

### Compatibility Testing
- Automated tests against Radarr v3 API
- Integration test suites
- Database migration testing

## Tools and Utilities

### Version Management Scripts
```bash
# Check current version
./radarr --version

# Validate version format
.github/scripts/validate-version.sh v1.2.3

# Generate changelog
.github/scripts/generate-changelog.sh v1.2.2..v1.2.3
```

### Docker Commands
```bash
# Pull specific version
docker pull ghcr.io/radarr/radarr-go:v1.2.3

# Pin to digest for production
docker pull ghcr.io/radarr/radarr-go@sha256:abc123...

# Check image metadata
docker inspect ghcr.io/radarr/radarr-go:latest
```

## Security Considerations

### Supply Chain Security
- SBOM (Software Bill of Materials) generation
- Provenance attestation for Docker images
- Vulnerability scanning with govulncheck
- Dependency security monitoring

### Release Verification
```bash
# Verify release signatures
gh release view v1.2.3 --json assets

# Verify Docker image digest
docker buildx imagetools inspect ghcr.io/radarr/radarr-go:v1.2.3

# Verify checksums
sha256sum -c radarr-v1.2.3-linux-amd64.tar.gz.sha256
```

## Communication Strategy

### Release Announcements
- GitHub Releases with detailed changelog
- Docker Hub descriptions
- Project documentation updates
- Community notifications

### Breaking Change Communication
- Advance notice in pre-release versions
- Migration guides
- Deprecation warnings
- Timeline for removal

## Best Practices

### For Developers
1. **Always validate versions** using provided scripts
2. **Test against multiple versions** of dependencies
3. **Document breaking changes** immediately
4. **Use semantic commit messages** for automatic changelog generation
5. **Update documentation** with version-specific changes

### For Users
1. **Pin to specific versions** in production
2. **Test pre-releases** in staging environments
3. **Subscribe to release notifications**
4. **Read changelogs** before upgrading
5. **Use digest pinning** for Docker deployments

### For Operators
1. **Monitor version adoption** metrics
2. **Plan upgrade windows** for major releases
3. **Maintain rollback capabilities**
4. **Test backup/restore** procedures with new versions
5. **Validate configurations** after upgrades

## Future Considerations

### Post-1.0 Evolution
- Long-term support (LTS) versions
- Extended support lifecycles
- Enterprise support tiers
- Feature flag systems

### Advanced Versioning
- Semantic import versioning for Go modules
- API versioning strategies
- Microservice version coordination
- Client library versioning

---

**Last Updated**: September 2025  
**Document Version**: 1.0  
**Applies to**: Radarr Go v0.9.0-alpha and later

For questions or suggestions regarding this versioning strategy, please create an issue in the [Radarr Go repository](https://github.com/radarr/radarr-go/issues).