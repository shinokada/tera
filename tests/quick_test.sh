#!/usr/bin/env bash

# Quick test script that runs a subset of critical tests
# Useful for rapid development feedback

set -e

cd "$(dirname "$0")"

echo "Running critical tests..."
echo

# Check if bats is installed
if ! command -v bats &> /dev/null; then
    echo "Error: BATS is not installed"
    exit 1
fi

# Run only the most important tests
echo "→ Testing menu structure..."
bats test_menu_structure.bats

echo
echo "→ Testing navigation..."
bats test_navigation.bats

echo
echo "✓ Critical tests passed!"
