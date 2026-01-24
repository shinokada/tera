#!/bin/bash
# Run Search Screen Tests

echo "================================"
echo "TERA Search Screen Test Suite"
echo "================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

echo "Running API Search Tests..."
echo "----------------------------"
if go test ./internal/api -v -run Search; then
    echo "✓ API Search Tests Passed"
    ((PASSED_TESTS++))
else
    echo "✗ API Search Tests Failed"
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))
echo ""

echo "Running UI Search Tests..."
echo "----------------------------"
if go test ./internal/ui -v -run Search; then
    echo "✓ UI Search Tests Passed"
    ((PASSED_TESTS++))
else
    echo "✗ UI Search Tests Failed"
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))
echo ""

echo "================================"
echo "Test Summary"
echo "================================"
echo "Total Test Suites: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo "All search tests passed! ✓"
    exit 0
else
    echo "Some tests failed. Please review the output above."
    exit 1
fi
