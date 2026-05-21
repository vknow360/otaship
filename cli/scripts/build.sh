#!/bin/bash

TARGET=$1
VERSION=$2
if [ -z "$VERSION" ]; then
    VERSION="dev"
fi

BUILD_DATE=$(date -u +"%Y-%m-%d")

LDFLAGS="-s -w -X github.com/vknow360/otaship/cli/internal/commands.Version=${VERSION} -X github.com/vknow360/otaship/cli/internal/commands.BuildDate=${BUILD_DATE}"

mkdir -p dist

case "${TARGET,,}" in
    "windows")
        echo "Building otaship-cli v${VERSION} for Windows..."
        GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/otaship-cli.exe cmd/otaship/main.go
        ;;
    "mac")
        echo "Building otaship-cli v${VERSION} for macOS (Apple Silicon)..."
        GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/otaship-cli_mac cmd/otaship/main.go
        ;;
    "linux"|*)
        echo "Building otaship-cli v${VERSION} for Linux..."
        GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/otaship-cli cmd/otaship/main.go
        ;;
esac

if [ $? -eq 0 ]; then
    echo "Build successful! [Version: $VERSION] [Date: $BUILD_DATE]"
    exit 0
else
    echo "Build failed!"
    exit 1
fi
