#!/bin/bash

# Cross-compilation script for PostgreSQL Client
# This script builds the application for multiple platforms
# Usage: ./build.sh [target_name]
#   target_name: e.g., Windows_x86_64, Linux_x86_64, macOS_ARM64, etc.
#   If omitted, builds for all platforms.

set -e

# get the project root directory, regardless of where the script is executed
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"
APP_NAME="postgresql-client"
VERSION=${VERSION:-"1.0.0"}
OUTPUT_DIR="${PROJECT_ROOT}/dist"

FILTER_TARGET="$1"

echo "Building $APP_NAME v$VERSION..."
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Define target platforms and architectures
TARGETS=(
    "darwin/amd64:macOS_x86_64"
    "darwin/arm64:macOS_ARM64"
    "linux/amd64:Linux_x86_64"
    "linux/386:Linux_i386"
    "linux/arm64:Linux_ARM64"
    "windows/amd64:Windows_x86_64.exe"
    "windows/386:Windows_i386.exe"
)

if [ -n "$FILTER_TARGET" ]; then
    # Remove .exe suffix for comparison, then re-add when matching
    echo "Filtering for target: $FILTER_TARGET"
    echo ""
fi

# Build for each target
for target in "${TARGETS[@]}"; do
    IFS=':' read -r GOOS_GOARCH OUTPUT_NAME <<< "$target"
    GOOS="${GOOS_GOARCH%/*}"
    GOARCH="${GOOS_GOARCH#*/}"

    # Apply filter: compare without .exe suffix
    if [ -n "$FILTER_TARGET" ]; then
        NAME_NO_EXT="${OUTPUT_NAME%.exe}"
        if [ "$NAME_NO_EXT" != "$FILTER_TARGET" ]; then
            continue
        fi
    fi

    echo "Building for $GOOS/$GOARCH -> $OUTPUT_NAME..."

    # Output file name with app name prefix
    OUTPUT_FILE="$OUTPUT_DIR/${APP_NAME}-${OUTPUT_NAME}"

    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -o "$OUTPUT_FILE" \
        -ldflags="-s -w" \
        "$PROJECT_ROOT"

    if [ $? -eq 0 ]; then
        echo "Built: $OUTPUT_NAME"
    else
        echo "Failed: $OUTPUT_NAME"
    fi
    echo ""
done

echo "Build completed! Output directory: $OUTPUT_DIR"
echo ""
echo "Files built:"
if command -v ls >/dev/null 2>&1; then
    ls -la "$OUTPUT_DIR" || echo "Directory listing not available"
elif command -v dir >/dev/null 2>&1; then
    dir "$OUTPUT_DIR" || echo "Directory listing not available"
else
    echo "Directory listing not available"
fi