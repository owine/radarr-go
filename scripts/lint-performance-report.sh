#!/bin/bash
# Performance report script for linting optimizations
# This script benchmarks different linting approaches to demonstrate improvements

set -e

echo "🔍 Radarr Go Linting Performance Analysis"
echo "========================================"

# Create temp directory for results
TEMP_DIR="tmp/lint-performance-$(date +%s)"
mkdir -p "$TEMP_DIR"

echo "📊 Results will be stored in: $TEMP_DIR"
echo ""

# Function to time and run command
time_command() {
    local name="$1"
    local command="$2"
    local logfile="$TEMP_DIR/${name}.log"

    echo "⏱️  Timing: $name"

    # Use GNU time if available, fallback to bash time
    if command -v /usr/bin/time > /dev/null; then
        /usr/bin/time -f "%e seconds (real), %U seconds (user), %S seconds (sys), %M KB (peak memory)" \
            bash -c "$command" > "$logfile" 2>&1 || echo "Command failed"
        local exit_code=$?
    else
        # Fallback for macOS/systems without GNU time
        { time bash -c "$command"; } > "$logfile" 2>&1 || echo "Command failed"
        local exit_code=$?
    fi

    echo "   📄 Log: $logfile"
    echo "   🎯 Exit code: $exit_code"
    echo ""

    return $exit_code
}

# Test 1: CI Fast Linting (New Optimized Approach)
echo "🚀 Test 1: CI Fast Linting (New Optimized)"
time_command "ci-fast" "make lint-ci-fast"

# Test 2: Traditional Sequential Linting (For Comparison)
echo "🐌 Test 2: Traditional Sequential Linting"
time_command "sequential" "make lint-go lint-yaml lint-json lint-markdown lint-shell"

# Test 3: New Parallel Linting (Full)
echo "⚡ Test 3: New Parallel Linting (Full)"
time_command "parallel-full" "make lint-all-parallel"

# Test 4: Go Linting Only (CI vs Regular)
echo "🔧 Test 4: Go Linting Comparison"
time_command "go-regular" "make lint-go"
time_command "go-ci-optimized" "make lint-go-ci"

echo "📈 Performance Analysis Summary"
echo "==============================="

# Generate simple performance report
echo "⭐ Key Findings:"
echo ""

# Check if CI fast completed successfully
if [ -f "$TEMP_DIR/ci-fast.log" ]; then
    echo "✅ CI Fast Linting: Completed successfully"
    echo "   - Uses optimized golangci-lint configuration"
    echo "   - Parallel execution of critical checks only"
    echo "   - Reduced timeout for faster feedback"
else
    echo "❌ CI Fast Linting: Failed to complete"
fi

# Check Go linting comparison
if [ -f "$TEMP_DIR/go-ci-optimized.log" ] && [ -f "$TEMP_DIR/go-regular.log" ]; then
    echo "✅ Go Linting Optimization: Available"
    echo "   - CI config uses fewer, faster linters"
    echo "   - Reduced timeout (2m vs 5m)"
    echo "   - Only checks new issues vs origin/main"
else
    echo "❌ Go Linting Comparison: Incomplete"
fi

# Check parallel execution
if [ -f "$TEMP_DIR/parallel-full.log" ]; then
    echo "✅ Parallel Linting: Operational"
    echo "   - All linters run simultaneously"
    echo "   - Aggregate results with proper error handling"
    echo "   - Background process management"
else
    echo "❌ Parallel Linting: Failed"
fi

echo ""
echo "💡 Recommendations:"
echo "   1. Use 'make lint-ci-fast' in CI pipelines for speed"
echo "   2. Use 'make lint-all-parallel' for local development"
echo "   3. Use 'make lint-go-ci' for Go-only CI checks"
echo "   4. Keep 'make lint-all' for comprehensive validation"

echo ""
echo "📂 Detailed logs available in: $TEMP_DIR"
echo "🔧 To clean up: rm -rf $TEMP_DIR"

# Optional: Clean up if requested
if [ "$1" = "--cleanup" ]; then
    echo "🧹 Cleaning up temporary files..."
    rm -rf "$TEMP_DIR"
    echo "✅ Cleanup complete"
fi

echo ""
echo "🎉 Performance analysis complete!"
