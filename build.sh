#!/usr/bin/env bash
set -euo pipefail

GO_BIN="$(command -v go || true)"
if [[ -z "${GO_BIN}" ]]; then
  if [[ -x "/usr/local/go/bin/go" ]]; then
    GO_BIN="/usr/local/go/bin/go"
  elif [[ -x "/c/Program Files/Go/bin/go.exe" ]]; then
    GO_BIN="/c/Program Files/Go/bin/go.exe"
  elif [[ -x "/mnt/c/Program Files/Go/bin/go.exe" ]]; then
    GO_BIN="/mnt/c/Program Files/Go/bin/go.exe"
  else
    echo "go not found on PATH" >&2
    exit 1
  fi
fi

"${GO_BIN}" version
"${GO_BIN}" env GOPATH GOROOT

VERSION="${1:-$(tr -d '[:space:]' < version.txt)}"
LDFLAGS="-X main.version=${VERSION} -s -w"

mkdir -p bin/linux/amd64 bin/linux/arm64 bin/darwin/amd64 bin/darwin/arm64 bin/windows/amd64

echo "Running local verification..."
"${GO_BIN}" test ./...
"${GO_BIN}" build ./...

echo "Building linux/amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 "${GO_BIN}" build -ldflags "${LDFLAGS}" -o bin/linux/amd64/azd-drasi .

echo "Building linux/arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 "${GO_BIN}" build -ldflags "${LDFLAGS}" -o bin/linux/arm64/azd-drasi .

echo "Building darwin/amd64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 "${GO_BIN}" build -ldflags "${LDFLAGS}" -o bin/darwin/amd64/azd-drasi .

echo "Building darwin/arm64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 "${GO_BIN}" build -ldflags "${LDFLAGS}" -o bin/darwin/arm64/azd-drasi .

echo "Building windows/amd64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 "${GO_BIN}" build -ldflags "${LDFLAGS}" -o bin/windows/amd64/azd-drasi.exe .

test -f bin/linux/amd64/azd-drasi
test -f bin/linux/arm64/azd-drasi
test -f bin/darwin/amd64/azd-drasi
test -f bin/darwin/arm64/azd-drasi
test -f bin/windows/amd64/azd-drasi.exe

echo "Done. Artifacts in bin/"
