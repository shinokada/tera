#!/bin/bash
# Test all the fixes

echo "Building tera..."
make clean
make build

if [ ! -f "./tera" ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✓ Build successful"
echo ""
echo "Testing fixes:"
echo "1. Station continues playing after quit - Press q while playing, verify station stops"
echo "2. Screen height - Check search menu shows all 6 options without scrolling"
echo "3. Save prompt after search play - Play station, press q, verify save prompt appears"
echo "4. Filter count - In search results, press / and type, verify count updates"
echo "5. Search menu height - Verify all menu items visible"
echo ""
echo "Press Enter to start testing..."
read

./tera
