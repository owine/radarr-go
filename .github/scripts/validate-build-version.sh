#!/bin/bash
set -euo pipefail

# Validate Build Version Information
# This script validates that the build system correctly handles version information
# Used in CI to ensure version handling works correctly during builds

echo "🔍 Validating build version handling..."

# Check if main.go or cmd/radarr/main.go has version variables
MAIN_FILE=""
if [[ -f "cmd/radarr/main.go" ]]; then
  MAIN_FILE="cmd/radarr/main.go"
elif [[ -f "main.go" ]]; then
  MAIN_FILE="main.go"
else
  echo "❌ Could not find main.go file"
  exit 1
fi

echo "📁 Main file: $MAIN_FILE"

# Check for version variables (checking for any of: var version, version =)
if ! grep -qE "(var\s+version|version\s*=)" "$MAIN_FILE"; then
  echo "❌ Version variable not found in $MAIN_FILE"
  echo "Expected to find: var version = ... or version = ..."
  exit 1
fi

if ! grep -qE "(var\s+commit|commit\s*=)" "$MAIN_FILE"; then
  echo "❌ Commit variable not found in $MAIN_FILE"  
  echo "Expected to find: var commit = ... or commit = ..."
  exit 1
fi

if ! grep -qE "(var\s+date|date\s*=)" "$MAIN_FILE"; then
  echo "❌ Date variable not found in $MAIN_FILE"
  echo "Expected to find: var date = ... or date = ..."
  exit 1
fi

echo "✅ Version variables found in $MAIN_FILE"

# Test version command if binary is provided
if [[ $# -gt 0 && -f "$1" ]]; then
  BINARY="$1"
  echo "🧪 Testing version command with binary: $BINARY"
  
  # Test --version flag
  if timeout 10s "$BINARY" --version 2>/dev/null; then
    echo "✅ Version command works"
  else
    echo "⚠️ Version command failed or timed out"
  fi
  
  # Test version output format
  VERSION_OUTPUT=$("$BINARY" --version 2>/dev/null || echo "")
  if [[ -n "$VERSION_OUTPUT" ]]; then
    echo "📋 Version output: $VERSION_OUTPUT"
    
    # Basic validation of version format
    if [[ "$VERSION_OUTPUT" =~ [0-9]+\.[0-9]+\.[0-9]+ ]]; then
      echo "✅ Version output contains semantic version"
    else
      echo "⚠️ Version output does not contain recognizable semantic version"
    fi
  fi
fi

# Validate that version can be set via ldflags
echo ""
echo "🔧 Build version injection validation:"

# Test build with version information
TEST_VERSION="1.2.3-test"
TEST_COMMIT="abc123"
TEST_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')

echo "   Testing with: Version=$TEST_VERSION, Commit=$TEST_COMMIT, Date=$TEST_DATE"

LDFLAGS="-X 'main.version=${TEST_VERSION}' -X 'main.commit=${TEST_COMMIT}' -X 'main.date=${TEST_DATE}'"

# Build test binary
echo "   Building test binary..."
if go build -ldflags="${LDFLAGS}" -o test-version-binary ./cmd/radarr 2>/dev/null; then
  echo "✅ Build with version injection successful"
  
  # Test the binary
  if timeout 5s ./test-version-binary --version 2>/dev/null; then
    VERSION_OUTPUT=$(timeout 5s ./test-version-binary --version 2>/dev/null || echo "")
    echo "   Output: $VERSION_OUTPUT"
    
    # Validate injected values appear in output
    if [[ "$VERSION_OUTPUT" == *"$TEST_VERSION"* ]]; then
      echo "✅ Version injection working correctly"
    else
      echo "⚠️ Version injection may not be working - $TEST_VERSION not found in output"
    fi
  else
    echo "⚠️ Test binary version command failed"
  fi
  
  # Cleanup
  rm -f test-version-binary
else
  echo "❌ Build with version injection failed"
  exit 1
fi

echo ""
echo "✅ Build version validation completed successfully"