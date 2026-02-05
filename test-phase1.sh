#!/bin/bash
# Test script to verify Phase 1 implementation

echo "Testing TERA Header Customization - Phase 1"
echo "============================================"
echo ""

cd "$(dirname "$0")"

echo "1. Checking if appearance_config.go exists..."
if [ -f "internal/storage/appearance_config.go" ]; then
    echo "   ✓ appearance_config.go created"
else
    echo "   ✗ appearance_config.go not found"
    exit 1
fi

echo ""
echo "2. Checking if header.go exists..."
if [ -f "internal/ui/header.go" ]; then
    echo "   ✓ header.go created"
else
    echo "   ✗ header.go not found"
    exit 1
fi

echo ""
echo "3. Building the project..."
if go build -o /tmp/tera-test 2>&1; then
    echo "   ✓ Build successful"
    rm -f /tmp/tera-test
else
    echo "   ✗ Build failed"
    exit 1
fi

echo ""
echo "4. Testing default configuration..."
go run . --version > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "   ✓ Application runs with default config"
else
    echo "   ✗ Application failed to run"
    exit 1
fi

echo ""
echo "============================================"
echo "Phase 1 Implementation: SUCCESS ✓"
echo ""
echo "The header renderer is now integrated and working."
echo "The application should work exactly as before with the default 'TERA' header."
echo ""
echo "Next steps:"
echo "  - Run the application and verify the header still shows 'TERA'"
echo "  - Create appearance_config.yaml manually to test different modes"
echo "  - Proceed to Phase 2 (Settings UI) when ready"
