#!/bin/bash
# Test the list management implementation

echo "Running tests for list management..."
go test -v ./internal/ui/... -run TestListManagement

echo ""
echo "Building application..."
go build -o tera ./cmd/tera

if [ $? -eq 0 ]; then
    echo "✓ Build successful!"
    echo ""
    echo "List Management features implemented:"
    echo "  1. Create new favorite lists"
    echo "  2. Delete lists (with My-favorites protection)"
    echo "  3. Rename lists (with My-favorites protection)"
    echo "  4. Show all lists"
    echo ""
    echo "Playback control fixes:"
    echo "  - Esc: Stop playback + show save prompt"
    echo "  - q: Exit application"
    echo "  - s: Save to Quick Favorites (during playback)"
    echo ""
    echo "Run './tera' to test!"
else
    echo "✗ Build failed"
    exit 1
fi
