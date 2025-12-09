#!/bin/bash
set -e

echo "Building QuietScan for all platforms..."

# Windows
echo "Building Windows (with icon)..."

# Clean old Windows resource files to ensure icon updates are picked up
rm -f cmd/quietscan/rsrc_windows_*.syso

# Generate Windows resources so Go embeds the icon
go-winres simply --arch amd64 --icon assets/icon.ico --manifest gui --out cmd/quietscan/rsrc

# Build Windows executable using the main package
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -mod=mod -ldflags "-H=windowsgui" \
  -o dist/quietscan-windows.exe ./cmd/quietscan

# macOS (Intel)
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -mod=mod -o QuietScan-macos-intel ./cmd/quietscan

# macOS (Apple Silicon)
echo "Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -mod=mod -o QuietScan-macos-arm64 ./cmd/quietscan

echo ""
echo "All builds successful!"
echo "Windows: quietscan-windows.exe"
echo "macOS Intel: QuietScan-macos-intel"
echo "macOS ARM64: QuietScan-macos-arm64"

