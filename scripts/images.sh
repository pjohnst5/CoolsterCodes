#!/bin/bash

# Image size validation script
# Scans for JPG/JPEG/PNG images larger than 1200x1200 pixels

set -e

echo "Scanning images for size violations..."

oversized=0
total=0

# Find and check all image files
for img in $(find content/ -type f \( -iname "*.jpg" -o -iname "*.jpeg" -o -iname "*.png" \) 2>/dev/null); do
    if [ -f "$img" ]; then
        total=$((total + 1))

        # Check if ImageMagick is available
        if ! command -v identify >/dev/null 2>&1; then
            echo "⚠️  ImageMagick not found."
            exit 1
        fi

        # Get image dimensions
        dimensions=$(identify -format "%wx%h" "$img" 2>/dev/null)
        if [ -n "$dimensions" ]; then
            width=${dimensions%x*}
            height=${dimensions#*x}

            # Check if image exceeds size limits
            if [ "$width" -gt 1200 ] || [ "$height" -gt 1200 ]; then
                echo "❌ OVERSIZED: $img ($dimensions)"
                oversized=$((oversized + 1))
            fi
        fi
    fi
done

if [ "$oversized" -gt 0 ]; then
    echo "❌ FAILED: Run 'make build' to optimize images"
    exit 1
else
    echo "✅ All images are properly sized"
fi
