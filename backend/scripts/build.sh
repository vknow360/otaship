#!/bin/bash

# $1 is the build target (windows, mac, linux)
TARGET=$1

# $2 is the optional version string (defaults to "dev" if empty)
VERSION=$2
if [ -z "$VERSION" ]; then
    VERSION="dev"
fi

BUILD_DATE=$(date -u +"%Y-%m-%d")

LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE}"

mkdir -p dist

# Evaluate the target platform
case "${TARGET,,}" in # ,, converts the string to lowercase automatically
    "windows")
        echo "Building for Windows..."
        GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/otaship-server.exe cmd/server/main.go
        ;;
    "mac")
        echo "Building for macOS (Apple Silicon)..."
        GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/otaship-server_mac cmd/server/main.go
        ;;
    "linux"|*)
        echo "Building for Linux..."
        GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/otaship-server cmd/server/main.go
        ;;
esac

# Check the compilation result
if [ $? -eq 0 ]; then
    echo "Build successful! [Version: $VERSION] [Date: $BUILD_DATE]"
    exit 0
else
    echo "Build failed!"
    exit 1
fi