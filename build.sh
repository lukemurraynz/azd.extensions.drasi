#!/usr/bin/env bash
set -euo pipefail

VERSION=$(cat version.txt | tr -d '[:space:]')
LDFLAGS="-X main.version=${VERSION} -s -w"

mkdir -p bin/linux/amd64 bin/linux/arm64 bin/darwin/amd64 bin/darwin/arm64 bin/windows/amd64

echo "Building linux/amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/linux/amd64/azd-drasi .

echo "Building linux/arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o bin/linux/arm64/azd-drasi .

echo "Building darwin/amd64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/darwin/amd64/azd-drasi .

echo "Building darwin/arm64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o bin/darwin/arm64/azd-drasi .

echo "Building windows/amd64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/windows/amd64/azd-drasi.exe .

echo "Done. Artifacts in bin/"
