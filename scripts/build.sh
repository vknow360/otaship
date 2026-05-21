#!/bin/bash

TARGET=$1
VERSION=$2

if [ -z "$TARGET" ]; then
    echo "Usage: ./build.sh <target> [version]"
    echo "Targets: windows, mac, linux"
    exit 1
fi

if [ -z "$VERSION" ]; then
    VERSION="dev"
fi

echo "========================================="
echo "Building Backend (Target: $TARGET, Version: $VERSION)"
echo "========================================="
cd backend
chmod +x scripts/build.sh
./scripts/build.sh "$TARGET" "$VERSION"
if [ $? -ne 0 ]; then
    echo "Backend build failed!"
    exit 1
fi
cd ..

echo ""
echo "========================================="
echo "Building CLI (Version: $VERSION)"
echo "========================================="
cd cli
chmod +x scripts/build.sh
./scripts/build.sh "$TARGET" "$VERSION"
if [ $? -ne 0 ]; then
    echo "CLI build failed!"
    exit 1
fi
cd ..

echo ""
echo "All builds completed successfully!"
