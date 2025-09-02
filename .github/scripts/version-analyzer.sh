#!/bin/bash
set -euo pipefail

# Enhanced Version Analysis and Strategy Implementation
# This script implements the automated versioning strategy from VERSIONING.md
# Usage: version-analyzer.sh VERSION [--json|--env|--docker-tags]

VERSION="${1:-}"
OUTPUT_FORMAT="${2:-env}"

if [[ -z "$VERSION" ]]; then
  echo "❌ Error: VERSION is required"
  echo "Usage: $0 VERSION [--json|--env|--docker-tags]"
  exit 1
fi

# Remove 'v' prefix if present
VERSION="${VERSION#v}"

# Validate semantic version format (strict)
if [[ ! $VERSION =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)(-[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*)?(\+[a-zA-Z0-9.-]+)?$ ]]; then
  echo "❌ Invalid semantic version format: $VERSION"
  echo "Expected format: X.Y.Z or X.Y.Z-prerelease or X.Y.Z-prerelease+build"
  echo "Examples: 1.0.0, 1.2.3-alpha, 1.2.3-beta.1, 1.0.0-rc.1+build.123"
  exit 1
fi

# Extract version components
MAJOR=$(echo "$VERSION" | sed -E 's/([0-9]+)\..*/\1/')
MINOR=$(echo "$VERSION" | sed -E 's/[0-9]+\.([0-9]+)\..*/\1/')
PATCH=$(echo "$VERSION" | sed -E 's/[0-9]+\.[0-9]+\.([0-9]+).*/\1/')

# Extract prerelease and build metadata
PRERELEASE=""
BUILD_META=""

if [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+-([^+]+)(\+(.+))?$ ]]; then
  PRERELEASE="${BASH_REMATCH[1]}"
  if [[ -n "${BASH_REMATCH[3]:-}" ]]; then
    BUILD_META="${BASH_REMATCH[3]}"
  fi
elif [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+\+(.+)$ ]]; then
  BUILD_META="${BASH_REMATCH[1]}"
fi

# Determine version phase and properties
IS_PRERELEASE=false
IS_PRE_1_0=false
RELEASE_STABILITY="stable"
PRERELEASE_TYPE=""

if [[ -n "$PRERELEASE" ]]; then
  IS_PRERELEASE=true

  # Determine prerelease type
  if [[ $PRERELEASE =~ ^alpha ]]; then
    PRERELEASE_TYPE="alpha"
    RELEASE_STABILITY="alpha"
  elif [[ $PRERELEASE =~ ^beta ]]; then
    PRERELEASE_TYPE="beta"
    RELEASE_STABILITY="beta"
  elif [[ $PRERELEASE =~ ^rc ]]; then
    PRERELEASE_TYPE="rc"
    RELEASE_STABILITY="rc"
  else
    PRERELEASE_TYPE="custom"
    RELEASE_STABILITY="prerelease"
  fi
fi

# Check if this is pre-1.0 (major version 0)
if [[ "$MAJOR" == "0" ]]; then
  IS_PRE_1_0=true
fi

# Determine release maturity
MATURITY_LEVEL="production"
if [[ "$IS_PRERELEASE" == "true" ]]; then
  case "$PRERELEASE_TYPE" in
    alpha)
      MATURITY_LEVEL="alpha"
      ;;
    beta)
      MATURITY_LEVEL="beta"
      ;;
    rc)
      MATURITY_LEVEL="release-candidate"
      ;;
    *)
      MATURITY_LEVEL="prerelease"
      ;;
  esac
elif [[ "$IS_PRE_1_0" == "true" ]]; then
  MATURITY_LEVEL="pre-production"
fi

# Generate Docker tags based on VERSIONING.md strategy
generate_docker_tags() {
  local base_image="${1:-ghcr.io/owner/repo}"
  local current_date=$(date +%Y.%m)

  # Always include version-specific tags
  local tags=(
    "v${VERSION}"
    "v${VERSION}-multi-db"
    "v${VERSION}-postgres"
    "v${VERSION}-mariadb"
  )

  if [[ "$IS_PRERELEASE" == "false" ]]; then
    # Production release tags (VERSIONING.md strategy)
    if [[ "$IS_PRE_1_0" == "false" ]]; then
      # Post-1.0 production releases get full tag set
      tags+=(
        "latest"
        "release"
        "stable"
        "stable-v${VERSION}"
        "${current_date}"
        "multi-db"
        "postgres"
        "mariadb"
      )
    else
      # Pre-1.0 production releases get limited tag set (no 'latest')
      tags+=(
        "release"
        "stable-v${VERSION}"
        "${current_date}"
      )
    fi

    # Special handling for 1.0.0 milestone
    if [[ "$VERSION" == "1.0.0" ]]; then
      tags+=(
        "milestone-v1"
        "production-ready"
      )
    fi
  else
    # Pre-release tags (VERSIONING.md strategy)
    tags+=(
      "testing"
      "testing-v${VERSION}"
      "prerelease"
    )

    # Type-specific prerelease tags
    case "$PRERELEASE_TYPE" in
      alpha)
        tags+=("alpha" "development")
        ;;
      beta)
        tags+=("beta" "testing-beta")
        ;;
      rc)
        tags+=("rc" "release-candidate")
        ;;
    esac
  fi

  # Convert array to comma-separated string
  local IFS=','
  echo "${tags[*]}"
}

# Version progression validation
validate_version_progression() {
  # This would ideally check against the latest Git tags
  # For now, provide basic validation rules

  local validation_result="valid"
  local warnings=()

  # Pre-1.0 specific validations
  if [[ "$IS_PRE_1_0" == "true" ]]; then
    # In pre-1.0, minor versions can introduce breaking changes
    if [[ "$PATCH" -gt 20 ]]; then
      warnings+=("High patch version ($PATCH) - consider bumping minor version")
    fi
  fi

  # Prerelease validation
  if [[ "$IS_PRERELEASE" == "true" ]]; then
    # Validate prerelease ordering (alpha -> beta -> rc)
    case "$PRERELEASE_TYPE" in
      alpha)
        # Alpha should have reasonable limits
        if [[ $PRERELEASE =~ alpha\.([0-9]+) ]] && [[ ${BASH_REMATCH[1]} -gt 10 ]]; then
          warnings+=("High alpha version (${BASH_REMATCH[1]}) - consider moving to beta")
        fi
        ;;
      beta)
        if [[ $PRERELEASE =~ beta\.([0-9]+) ]] && [[ ${BASH_REMATCH[1]} -gt 5 ]]; then
          warnings+=("High beta version (${BASH_REMATCH[1]}) - consider moving to RC")
        fi
        ;;
    esac
  fi

  echo "$validation_result"
  if [[ ${#warnings[@]} -gt 0 ]]; then
    printf "Warning: %s\n" "${warnings[@]}" >&2
  fi
}

# Generate version compatibility information
generate_compatibility_info() {
  local api_compatibility="radarr-v3"
  local go_module_compatibility="compatible"
  local breaking_changes="none"

  if [[ "$IS_PRE_1_0" == "true" ]]; then
    go_module_compatibility="pre-1.0-unstable"
    if [[ "$PRERELEASE_TYPE" == "alpha" ]]; then
      breaking_changes="possible"
    fi
  fi

  echo "api_compatibility=${api_compatibility}"
  echo "go_module_compatibility=${go_module_compatibility}"
  echo "breaking_changes=${breaking_changes}"
}

# Output results based on format
case "$OUTPUT_FORMAT" in
  --json)
    cat <<EOF
{
  "version": "${VERSION}",
  "major": ${MAJOR},
  "minor": ${MINOR},
  "patch": ${PATCH},
  "prerelease": "${PRERELEASE}",
  "build_metadata": "${BUILD_META}",
  "is_prerelease": ${IS_PRERELEASE},
  "is_pre_1_0": ${IS_PRE_1_0},
  "prerelease_type": "${PRERELEASE_TYPE}",
  "release_stability": "${RELEASE_STABILITY}",
  "maturity_level": "${MATURITY_LEVEL}",
  "validation_result": "$(validate_version_progression)"
}
EOF
    ;;

  --docker-tags)
    generate_docker_tags "${3:-ghcr.io/owner/repo}"
    ;;

  --env|*)
    # GitHub Actions environment format
    echo "VERSION=${VERSION}"
    echo "VERSION_MAJOR=${MAJOR}"
    echo "VERSION_MINOR=${MINOR}"
    echo "VERSION_PATCH=${PATCH}"
    echo "VERSION_PRERELEASE=${PRERELEASE}"
    echo "VERSION_BUILD_META=${BUILD_META}"
    echo "IS_PRERELEASE=${IS_PRERELEASE}"
    echo "IS_PRE_1_0=${IS_PRE_1_0}"
    echo "PRERELEASE_TYPE=${PRERELEASE_TYPE}"
    echo "RELEASE_STABILITY=${RELEASE_STABILITY}"
    echo "MATURITY_LEVEL=${MATURITY_LEVEL}"
    echo "VALIDATION_RESULT=$(validate_version_progression)"
    generate_compatibility_info
    ;;
esac

# Summary output to stderr for GitHub Actions logs
{
  echo "🔍 Version Analysis Results:"
  echo "   Version: ${VERSION}"
  echo "   Components: Major=${MAJOR}, Minor=${MINOR}, Patch=${PATCH}"
  if [[ -n "$PRERELEASE" ]]; then
    echo "   Prerelease: ${PRERELEASE} (${PRERELEASE_TYPE})"
  fi
  if [[ -n "$BUILD_META" ]]; then
    echo "   Build: ${BUILD_META}"
  fi
  echo "   Pre-1.0: ${IS_PRE_1_0}"
  echo "   Prerelease: ${IS_PRERELEASE}"
  echo "   Stability: ${RELEASE_STABILITY}"
  echo "   Maturity: ${MATURITY_LEVEL}"
  echo "   Status: $(validate_version_progression)"
} >&2
