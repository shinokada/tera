#!/bin/bash
# Fix duplicate file issue
rm -f internal/ui/blocklist_enhanced.go
echo "Removed duplicate file: internal/ui/blocklist_enhanced.go"
echo "Now you can run: make clean-all && make lint && make build"
