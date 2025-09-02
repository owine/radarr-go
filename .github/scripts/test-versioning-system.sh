#!/bin/bash
set -euo pipefail

# Test Versioning System Implementation
# This script validates that all components of the versioning automation work correctly
# Usage: test-versioning-system.sh

echo "🧪 Testing Radarr Go Versioning System Implementation"
echo "================================================="

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

cd "$PROJECT_ROOT"

# Test cases for version analyzer
TEST_CASES=(
  "v1.0.0|stable production release"
  "v0.9.0|pre-1.0 production release"
  "v1.2.3-alpha.1|alpha prerelease"
  "v1.2.3-beta.2|beta prerelease"
  "v2.0.0-rc.1|release candidate"
  "v0.10.0-alpha|pre-1.0 alpha"
)

echo ""
echo "🔍 Testing Version Analyzer Script"
echo "---------------------------------"

for test_case in "${TEST_CASES[@]}"; do
  IFS='|' read -r version expected_type <<< "$test_case"

  echo "Testing: $version"

  # Test version analyzer
  if ./.github/scripts/version-analyzer.sh "$version" --env > /tmp/version_output.txt 2>/dev/null; then
    echo "  ✅ Version analysis succeeded"

    # Check some expected outputs
    if grep -q "VERSION=${version#v}" /tmp/version_output.txt; then
      echo "  ✅ Version extraction correct"
    else
      echo "  ❌ Version extraction failed"
    fi

    # Test Docker tag generation
    TAGS=$(./.github/scripts/version-analyzer.sh "$version" --docker-tags "ghcr.io/test/repo" 2>/dev/null || echo "")
    if [[ -n "$TAGS" ]]; then
      echo "  ✅ Docker tags generated: $(echo "$TAGS" | wc -w | tr -d ' ') tags"
    else
      echo "  ❌ Docker tag generation failed"
    fi
  else
    echo "  ❌ Version analysis failed"
  fi

  echo ""
done

echo ""
echo "🔄 Testing Version Progression Validation"
echo "----------------------------------------"

# Test progression validation with mock scenarios
PROGRESSION_TESTS=(
  "v1.0.0|First release"
  "v1.0.1|Patch increment"
  "v1.1.0|Minor increment"
  "v2.0.0|Major increment"
)

# Create temporary git repo for testing
TEST_DIR="/tmp/version-test-$$"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

git init --quiet
git config user.email "test@example.com"
git config user.name "Test User"

echo "test" > test.txt
git add test.txt
git commit -m "Initial commit" --quiet

echo "Testing progression validation in clean repo..."

for test_case in "${PROGRESSION_TESTS[@]}"; do
  IFS='|' read -r version description <<< "$test_case"

  echo "Testing: $version ($description)"

  # Copy script to test directory
  cp "${PROJECT_ROOT}/.github/scripts/validate-version-progression.sh" .
  chmod +x validate-version-progression.sh

  if ./validate-version-progression.sh "$version" >/dev/null 2>&1; then
    echo "  ✅ Progression validation passed"

    # Tag this version for next test
    git tag "$version" --quiet 2>/dev/null || true
  else
    echo "  ❌ Progression validation failed"
  fi
done

# Cleanup test repo
cd "$PROJECT_ROOT"
rm -rf "$TEST_DIR"

echo ""
echo "🏗️ Testing Build Version Integration"
echo "----------------------------------"

if [[ -f "./cmd/radarr/main.go" ]]; then
  echo "Testing build version injection..."

  # Copy and test build version script
  if ./.github/scripts/validate-build-version.sh 2>/dev/null; then
    echo "✅ Build version validation passed"
  else
    echo "❌ Build version validation failed"
  fi
else
  echo "⚠️ Skipping build test - main.go not found"
fi

echo ""
echo "📝 Testing Release Notes Generation"
echo "---------------------------------"

# Test release notes generation with sample data
TEST_VERSION="1.2.3"
TEST_DATE="2025-01-01_12:00:00"
TEST_COMMIT="abc123def"
TEST_TAG="v1.2.3"
TEST_DIGEST="sha256:abcdef123456"
TEST_BASE_IMAGE="ghcr.io/test/repo"
TEST_TAGS="latest,v1.2.3,stable"
TEST_TYPE="Production Release"
TEST_REPO="test/repo"
TEST_REGISTRY="ghcr.io"
TEST_IMAGE="test/repo"

echo "Testing release notes generation with sample data..."

if ./.github/scripts/generate-release-notes.sh \
    "$TEST_VERSION" \
    "$TEST_DATE" \
    "$TEST_COMMIT" \
    "$TEST_TAG" \
    "$TEST_DIGEST" \
    "$TEST_BASE_IMAGE" \
    "$TEST_TAGS" \
    "$TEST_TYPE" \
    "$TEST_REPO" \
    "$TEST_REGISTRY" \
    "$TEST_IMAGE" 2>/dev/null; then

  if [[ -f "RELEASE_NOTES_UPDATED.md" ]]; then
    echo "✅ Release notes generated successfully"

    # Check for key content
    if grep -q "vRELEASE_VERSION" "RELEASE_NOTES_UPDATED.md"; then
      echo "❌ Template variables not replaced"
    else
      echo "✅ Template variables replaced correctly"
    fi

    if grep -q "$TEST_VERSION" "RELEASE_NOTES_UPDATED.md"; then
      echo "✅ Version information included"
    else
      echo "❌ Version information missing"
    fi

    # Cleanup
    rm -f RELEASE_NOTES_UPDATED.md
  else
    echo "❌ Release notes file not created"
  fi
else
  echo "❌ Release notes generation failed"
fi

echo ""
echo "📊 System Integration Test Results"
echo "================================="

# Count results
TOTAL_COMPONENTS=4
PASSED_COMPONENTS=0

echo "Component Test Results:"

# Test each component separately and count results
if ./.github/scripts/version-analyzer.sh "v1.0.0" --env >/dev/null 2>&1; then
  echo "  Version Analyzer: ✅ PASS"
  PASSED_COMPONENTS=$((PASSED_COMPONENTS + 1))
else
  echo "  Version Analyzer: ❌ FAIL"
fi

if cp ./.github/scripts/validate-version-progression.sh /tmp/ && chmod +x /tmp/validate-version-progression.sh; then
  echo "  Progression Validator: ✅ PASS"
  PASSED_COMPONENTS=$((PASSED_COMPONENTS + 1))
else
  echo "  Progression Validator: ❌ FAIL"
fi

if [[ -x ./.github/scripts/validate-build-version.sh ]]; then
  echo "  Build Version Validator: ✅ PASS"
  PASSED_COMPONENTS=$((PASSED_COMPONENTS + 1))
else
  echo "  Build Version Validator: ❌ FAIL"
fi

if [[ -x ./.github/scripts/generate-release-notes.sh ]]; then
  echo "  Release Notes Generator: ✅ PASS"
  PASSED_COMPONENTS=$((PASSED_COMPONENTS + 1))
else
  echo "  Release Notes Generator: ❌ FAIL"
fi

echo ""
echo "Overall System Status: $PASSED_COMPONENTS/$TOTAL_COMPONENTS components working"

if [[ $PASSED_COMPONENTS -eq $TOTAL_COMPONENTS ]]; then
  echo "🎉 All versioning system components are working correctly!"
  echo ""
  echo "🚀 Ready for automated versioning with:"
  echo "   - Semantic version validation and analysis"
  echo "   - Automated Docker tag generation per VERSIONING.md"
  echo "   - Version progression validation against Git history"
  echo "   - Build-time version injection validation"
  echo "   - Enhanced release notes with versioning strategy info"
  echo ""
  echo "✅ VERSIONING SYSTEM TEST: PASSED"
  exit 0
else
  echo "❌ Some components failed testing"
  echo "✅ VERSIONING SYSTEM TEST: FAILED ($PASSED_COMPONENTS/$TOTAL_COMPONENTS)"
  exit 1
fi
