#!/bin/bash
# Build and verify the application

set -e

echo "Building TERA..."
go build -o tera ./cmd/tera

echo "✓ Build successful!"
echo
echo "You can now test the application with: ./tera"
echo
echo "New features:"
echo "  1. Fixed playback controls: 'esc' stops and shows save prompt, 'q' exits"
echo "  2. Save prompt after stopping playback"
echo "  3. List Management Menu implemented with:"
echo "     - Create new lists"
echo "     - Delete lists (with protection for My-favorites)"
echo "     - Rename lists"
echo "     - Show all lists"
echo
echo "Test the List Management menu:"
echo "  1. Run ./tera"
echo "  2. Select '3) Manage Lists'"
echo "  3. Try creating, renaming, and deleting lists"
