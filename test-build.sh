#!/bin/bash
# Test build script

cd /Users/shinichiokada/Terminal-Tools/tera

echo "=== Building TERA ==="
make clean && make build

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Build successful!"
    echo ""
    echo "To test:"
    echo "  ./tera"
    echo ""
    echo "Navigate to: Main Menu > Settings (6) > Appearance (2)"
else
    echo ""
    echo "❌ Build failed. Check errors above."
    exit 1
fi
