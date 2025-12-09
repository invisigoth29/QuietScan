#!/bin/bash
# Clean build script - purges ALL caches before building
# Use this when the Windows icon doesn't update

set -e

echo "====================================="
echo "QuietScan Clean Build"
echo "Purging all caches and rebuilding..."
echo "====================================="
echo ""

# Clean Windows resource files
echo "1. Cleaning Windows resource files (.syso)..."
rm -f cmd/quietscan/rsrc_windows_*.syso
rm -f cmd/quietscan/*.syso

# Clean built executables
echo "2. Cleaning built executables..."
rm -f dist/quietscan-windows.exe
rm -f QuietScan-macos-intel
rm -f QuietScan-macos-arm64

# Clean Go build cache (optional but thorough)
echo "3. Cleaning Go build cache..."
go clean -cache -modcache -i -r 2>/dev/null || true

# Clean any go-winres cache
echo "4. Cleaning go-winres cache..."
rm -rf ~/.cache/go-winres 2>/dev/null || true

echo ""
echo "Cache purge complete!"
echo ""
echo "====================================="
echo "Building fresh..."
echo "====================================="
echo ""

# Run the normal build
./build-all.sh

echo ""
echo "====================================="
echo "Clean build complete!"
echo "====================================="
echo ""
echo "If the Windows icon still doesn't update:"
echo "1. Copy dist/quietscan-windows.exe to a Windows machine"
echo "2. Run clear-windows-icon-cache.bat on Windows"
echo "3. Check the icon again"
echo ""
