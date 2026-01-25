#!/bin/bash
# Test script for arrow key navigation implementation

echo "Testing menu component..."
go test -v ./internal/ui/components/

echo ""
echo "Building application..."
go build -o tera ./cmd/tera/

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Build successful!"
    echo ""
    echo "Run './tera' to test the application"
    echo ""
    echo "Test the following:"
    echo "  1. Main menu arrow key navigation (↑↓)"
    echo "  2. Main menu vim key navigation (j/k)"
    echo "  3. Main menu number shortcuts (1-6)"
    echo "  4. Search menu arrow key navigation"
    echo "  5. Search menu vim key navigation"
    echo "  6. Search menu number shortcuts (1-6)"
    echo "  7. Esc/0 navigation"
else
    echo ""
    echo "✗ Build failed"
    exit 1
fi
