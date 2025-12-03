#!/bin/bash
# Create a multi-size ICO file for Windows
# This script uses ImageMagick if available, otherwise provides instructions

INPUT="assets/icon.png"
OUTPUT="assets/icon.ico"

if [ ! -f "$INPUT" ]; then
    echo "Error: $INPUT not found!"
    exit 1
fi

if command -v magick &> /dev/null; then
    echo "Creating multi-size ICO file using ImageMagick..."
    magick convert "$INPUT" -define icon:auto-resize=256,128,64,48,32,16 "$OUTPUT"
    echo "✓ Created $OUTPUT with multiple sizes"
elif command -v convert &> /dev/null; then
    echo "Creating multi-size ICO file using ImageMagick (legacy)..."
    convert "$INPUT" -define icon:auto-resize=256,128,64,48,32,16 "$OUTPUT"
    echo "✓ Created $OUTPUT with multiple sizes"
else
    echo "ImageMagick not found. Please use an online converter:"
    echo ""
    echo "1. Go to: https://convertio.co/png-ico/"
    echo "2. Upload: $INPUT"
    echo "3. Enable 'Multiple sizes' option"
    echo "4. Download and save as: $OUTPUT"
    echo ""
    echo "Or install ImageMagick:"
    echo "  macOS: brew install imagemagick"
    echo "  Windows: Download from https://imagemagick.org/"
    exit 1
fi


