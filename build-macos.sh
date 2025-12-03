#!/bin/bash
set -e

echo "Building QuietScan for macOS..."

export GOOS=darwin
export GOARCH=amd64
export CGO_ENABLED=1

go build -mod=mod -o QuietScan ./cmd/quietscan

echo ""
echo "Build successful: QuietScan"
echo "This is a fully self-contained executable - no external files needed!"


