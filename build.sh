#!/bin/bash

set -e

echo "Building NEMS..."

# Clean previous builds
rm -rf build
mkdir -p build/dist

echo "1/3 Building Vue Frontend..."
cd frontend
npm install
npm run build
cp -r dist/* ../build/dist/
cd ..

echo "2/3 Compiling Go Backend for linux/arm64..."
cd backend
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.BuildNumber=$(git describe --tags --always)" -o ../build/nems-server *.go
cd ..

echo "3/3 Bundling release archive..."
cd build
# Create archive containing the executable and the dist folder
tar -czvf nems-release-arm64.tar.gz nems-server dist
cd ..

echo "Build complete! Output is located at build/nems-release-arm64.tar.gz"
