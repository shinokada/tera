#!/usr/bin/env bash

# Test runner script for TERA
# This script runs all tests and provides a summary

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "        TERA Test Suite Runner"
echo "========================================="
echo

# Check if bats is installed
if ! command -v bats &> /dev/null; then
    echo -e "${RED}Error: BATS is not installed${NC}"
    echo "Please install BATS first:"
    echo "  macOS: brew install bats-core"
    echo "  Linux: sudo apt-get install bats"
    echo "  Manual: https://github.com/bats-core/bats-core"
    exit 1
fi

echo -e "${GREEN}✓ BATS is installed${NC}"
echo

# Change to tests directory
cd "$(dirname "$0")"

echo "Running all tests..."
echo

# Run tests
if bats .; then
    echo
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}   All tests passed! ✓${NC}"
    echo -e "${GREEN}=========================================${NC}"
    exit 0
else
    echo
    echo -e "${RED}=========================================${NC}"
    echo -e "${RED}   Some tests failed! ✗${NC}"
    echo -e "${RED}=========================================${NC}"
    exit 1
fi
