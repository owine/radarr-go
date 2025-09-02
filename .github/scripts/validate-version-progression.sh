#!/bin/bash
set -euo pipefail

# Validate Version Progression Against Git History
# This script ensures that new versions follow proper semantic versioning progression
# Usage: validate-version-progression.sh VERSION

VERSION="${1:-}"

if [[ -z "$VERSION" ]]; then
  echo "❌ Error: VERSION is required"
  echo "Usage: $0 VERSION"
  exit 1
fi

# Remove 'v' prefix if present
VERSION="${VERSION#v}"

echo "🔍 Validating version progression for: v$VERSION"

# Get all existing tags and extract version numbers
echo "📋 Fetching existing version tags..."
# Exclude the current version being validated from the comparison
EXISTING_TAGS=$(git tag -l "v*" | grep -E "^v[0-9]+\.[0-9]+\.[0-9]+" | grep -v "^v${VERSION}$" | sort -V || echo "")

if [[ -z "$EXISTING_TAGS" ]]; then
  echo "ℹ️ No existing version tags found - this appears to be the first release"
  echo "✅ Version progression validation: PASSED (first release)"
  exit 0
fi

echo "📊 Found existing tags:"
echo "$EXISTING_TAGS" | sed 's/^/   - /'

# Get the latest stable and prerelease versions
LATEST_STABLE=$(echo "$EXISTING_TAGS" | grep -v -E "-(alpha|beta|rc)" | tail -1 | sed 's/^v//' || echo "")
LATEST_ANY=$(echo "$EXISTING_TAGS" | tail -1 | sed 's/^v//' || echo "")

echo ""
echo "📈 Version Analysis:"
echo "   Latest stable: ${LATEST_STABLE:-none}"
echo "   Latest any: ${LATEST_ANY:-none}"
echo "   New version: $VERSION"

# Parse version components for comparison
parse_version() {
  local ver="$1"
  echo "$ver" | sed -E 's/([0-9]+)\.([0-9]+)\.([0-9]+).*/\1 \2 \3/'
}

# Compare two semantic versions (without prerelease info)
# Returns: -1 (first < second), 0 (equal), 1 (first > second)
compare_versions() {
  local ver1="$1"
  local ver2="$2"

  # Remove prerelease suffixes for base version comparison
  local base1=$(echo "$ver1" | sed 's/-.*$//')
  local base2=$(echo "$ver2" | sed 's/-.*$//')

  read -r maj1 min1 pat1 <<< "$(parse_version "$base1")"
  read -r maj2 min2 pat2 <<< "$(parse_version "$base2")"

  if [[ $maj1 -gt $maj2 ]]; then echo 1; return; fi
  if [[ $maj1 -lt $maj2 ]]; then echo -1; return; fi

  if [[ $min1 -gt $min2 ]]; then echo 1; return; fi
  if [[ $min1 -lt $min2 ]]; then echo -1; return; fi

  if [[ $pat1 -gt $pat2 ]]; then echo 1; return; fi
  if [[ $pat1 -lt $pat2 ]]; then echo -1; return; fi

  echo 0
}

# Validation logic
VALIDATION_ERRORS=()
VALIDATION_WARNINGS=()

# Check if new version is greater than latest
if [[ -n "$LATEST_ANY" ]]; then
  COMPARISON=$(compare_versions "$VERSION" "$LATEST_ANY")

  if [[ $COMPARISON -lt 0 ]]; then
    VALIDATION_ERRORS+=("Version $VERSION is older than existing version $LATEST_ANY")
  elif [[ $COMPARISON -eq 0 ]]; then
    # Same base version - check prerelease progression
    NEW_PRERELEASE=""
    LATEST_PRERELEASE=""

    if [[ $VERSION =~ -(.+)$ ]]; then
      NEW_PRERELEASE="${BASH_REMATCH[1]}"
    fi

    if [[ $LATEST_ANY =~ -(.+)$ ]]; then
      LATEST_PRERELEASE="${BASH_REMATCH[1]}"
    fi

    # Both have same base version
    if [[ -n "$NEW_PRERELEASE" && -n "$LATEST_PRERELEASE" ]]; then
      # Both are prereleases - validate progression (alpha < beta < rc)
      case "$LATEST_PRERELEASE" in
        alpha*)
          if [[ $NEW_PRERELEASE =~ ^alpha ]]; then
            # Compare alpha numbers
            LATEST_NUM=$(echo "$LATEST_PRERELEASE" | sed 's/alpha\.*//' | sed 's/.*\([0-9][0-9]*\)$/\1/' || echo "0")
            NEW_NUM=$(echo "$NEW_PRERELEASE" | sed 's/alpha\.*//' | sed 's/.*\([0-9][0-9]*\)$/\1/' || echo "0")
            if [[ ${NEW_NUM:-0} -le ${LATEST_NUM:-0} ]]; then
              VALIDATION_ERRORS+=("Alpha version $NEW_PRERELEASE is not greater than existing $LATEST_PRERELEASE")
            fi
          fi
          # beta or rc after alpha is OK
          ;;
        beta*)
          if [[ $NEW_PRERELEASE =~ ^alpha ]]; then
            VALIDATION_ERRORS+=("Cannot release alpha after beta")
          elif [[ $NEW_PRERELEASE =~ ^beta ]]; then
            LATEST_NUM=$(echo "$LATEST_PRERELEASE" | sed 's/beta\.*//' | sed 's/.*\([0-9][0-9]*\)$/\1/' || echo "0")
            NEW_NUM=$(echo "$NEW_PRERELEASE" | sed 's/beta\.*//' | sed 's/.*\([0-9][0-9]*\)$/\1/' || echo "0")
            if [[ ${NEW_NUM:-0} -le ${LATEST_NUM:-0} ]]; then
              VALIDATION_ERRORS+=("Beta version $NEW_PRERELEASE is not greater than existing $LATEST_PRERELEASE")
            fi
          fi
          # rc after beta is OK
          ;;
        rc*)
          if [[ $NEW_PRERELEASE =~ ^(alpha|beta) ]]; then
            VALIDATION_ERRORS+=("Cannot release alpha/beta after rc")
          elif [[ $NEW_PRERELEASE =~ ^rc ]]; then
            LATEST_NUM=$(echo "$LATEST_PRERELEASE" | sed 's/rc\.*//' | sed 's/.*\([0-9][0-9]*\)$/\1/' || echo "0")
            NEW_NUM=$(echo "$NEW_PRERELEASE" | sed 's/rc\.*//' | sed 's/.*\([0-9][0-9]*\)$/\1/' || echo "0")
            if [[ ${NEW_NUM:-0} -le ${LATEST_NUM:-0} ]]; then
              VALIDATION_ERRORS+=("RC version $NEW_PRERELEASE is not greater than existing $LATEST_PRERELEASE")
            fi
          fi
          ;;
      esac
    elif [[ -n "$NEW_PRERELEASE" && -z "$LATEST_PRERELEASE" ]]; then
      VALIDATION_ERRORS+=("Cannot release prerelease $VERSION after stable release $LATEST_ANY")
    elif [[ -z "$NEW_PRERELEASE" && -n "$LATEST_PRERELEASE" ]]; then
      # Releasing stable after prerelease is OK
      echo "✅ Releasing stable version after prerelease"
    else
      # Same version, both stable
      VALIDATION_ERRORS+=("Version $VERSION already exists")
    fi
  fi
fi

# Check version increment rules
if [[ -n "$LATEST_STABLE" ]]; then
  # Parse versions
  read -r latest_maj latest_min latest_pat <<< "$(parse_version "$LATEST_STABLE")"
  read -r new_maj new_min new_pat <<< "$(parse_version "$VERSION")"

  # Remove prerelease for increment validation
  VERSION_BASE=$(echo "$VERSION" | sed 's/-.*$//')
  read -r new_maj new_min new_pat <<< "$(parse_version "$VERSION_BASE")"

  echo ""
  echo "🔢 Version increment analysis:"
  echo "   Previous: $latest_maj.$latest_min.$latest_pat"
  echo "   New base: $new_maj.$new_min.$new_pat"

  # Major version increment
  if [[ $new_maj -gt $latest_maj ]]; then
    if [[ $new_min -ne 0 || $new_pat -ne 0 ]]; then
      VALIDATION_WARNINGS+=("Major version bump should reset minor and patch to 0")
    fi
    echo "   → Major version increment detected"
  # Minor version increment
  elif [[ $new_maj -eq $latest_maj && $new_min -gt $latest_min ]]; then
    if [[ $new_pat -ne 0 ]]; then
      VALIDATION_WARNINGS+=("Minor version bump should reset patch to 0")
    fi
    echo "   → Minor version increment detected"
  # Patch version increment
  elif [[ $new_maj -eq $latest_maj && $new_min -eq $latest_min && $new_pat -gt $latest_pat ]]; then
    echo "   → Patch version increment detected"
  # Pre-1.0 special handling
  elif [[ $new_maj -eq 0 ]]; then
    echo "   → Pre-1.0 version (flexible increment rules apply)"
    # In pre-1.0, breaking changes can happen in minor versions
    if [[ $new_min -lt $latest_min ]]; then
      VALIDATION_ERRORS+=("Minor version cannot decrease in pre-1.0")
    fi
  else
    VALIDATION_ERRORS+=("Invalid version increment: $LATEST_STABLE → $VERSION_BASE")
  fi
fi

# Special milestone validations
if [[ "$VERSION" == "1.0.0" ]]; then
  echo ""
  echo "🎉 Major milestone detected: v1.0.0"
  echo "   This represents the first production-ready release"
  echo "   API stability promise begins with this version"
fi

# Check for potential issues
CURRENT_MAJOR=$(echo "$VERSION" | sed -E 's/([0-9]+)\..*/\1/')
if [[ $CURRENT_MAJOR -eq 0 ]]; then
  echo ""
  echo "⚠️  Pre-1.0 Version Notes:"
  echo "   - Breaking changes allowed in minor versions"
  echo "   - API stability not guaranteed"
  echo "   - Production use requires careful consideration"
fi

# Output results
echo ""
if [[ ${#VALIDATION_ERRORS[@]} -gt 0 ]]; then
  echo "❌ Validation FAILED:"
  for error in "${VALIDATION_ERRORS[@]}"; do
    echo "   ❌ $error"
  done
  exit 1
fi

if [[ ${#VALIDATION_WARNINGS[@]} -gt 0 ]]; then
  echo "⚠️  Validation WARNINGS:"
  for warning in "${VALIDATION_WARNINGS[@]}"; do
    echo "   ⚠️  $warning"
  done
fi

echo "✅ Version progression validation: PASSED"

# Additional context for GitHub Actions
echo ""
echo "📊 Context for release:"
if [[ -n "$LATEST_STABLE" ]]; then
  echo "   Upgrading from: v$LATEST_STABLE"
else
  echo "   This is the first stable release"
fi

if [[ $VERSION =~ -(.+)$ ]]; then
  PRERELEASE_TYPE="${BASH_REMATCH[1]}"
  echo "   Release type: Prerelease ($PRERELEASE_TYPE)"
  echo "   Production ready: No"
else
  echo "   Release type: Stable"
  if [[ $CURRENT_MAJOR -eq 0 ]]; then
    echo "   Production ready: Limited (pre-1.0)"
  else
    echo "   Production ready: Yes"
  fi
fi

echo ""
echo "✅ Version v$VERSION is valid for release"
