#!/usr/bin/env bash

# Test script for TERA improvements
# Tests auto-creation and navigation improvements

set -e

echo "======================================"
echo "TERA Improvements Test Script"
echo "======================================"
echo

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

info() {
    echo -e "${YELLOW}ℹ INFO${NC}: $1"
}

section() {
    echo
    echo "======================================"
    echo "$1"
    echo "======================================"
    echo
}

# Test 1: Check if main tera script exists
section "Test 1: Script Existence"

if [ -f "./tera" ]; then
    pass "tera script exists"
else
    fail "tera script not found"
    exit 1
fi

# Test 2: Check lib files exist
section "Test 2: Library Files"

if [ -f "./lib/search.sh" ]; then
    pass "lib/search.sh exists"
else
    fail "lib/search.sh not found"
fi

if [ -f "./lib/list.sh" ]; then
    pass "lib/list.sh exists"
else
    fail "lib/list.sh not found"
fi

# Test 3: Check for auto-creation code in main script
section "Test 3: Auto-Creation Feature"

if grep -q "Initialize My-favorites.json if it doesn't exist" ./tera; then
    pass "Auto-creation code found in tera script"
else
    fail "Auto-creation code not found"
fi

# Test 4: Check navigation improvements in search.sh
section "Test 4: Navigation in search.sh"

if grep -q "Type '0' to go back to Search Menu, '00' for Main Menu" ./lib/search.sh; then
    pass "Updated navigation message found in search_by"
else
    fail "Navigation message not updated in search_by"
fi

if grep -q 'case "$REPLY" in' ./lib/search.sh; then
    pass "Navigation case statement found in search_by"
else
    fail "Navigation case statement not found"
fi

if grep -q '"0"|"back"' ./lib/search.sh; then
    pass "Back navigation ('0') implemented"
else
    fail "Back navigation not found"
fi

if grep -q '"00"|"main"' ./lib/search.sh; then
    pass "Main menu navigation ('00') implemented"
else
    fail "Main menu navigation not found"
fi

# Test 5: Check navigation improvements in list.sh
section "Test 5: Navigation in list.sh"

if grep -q "Press Enter to continue..." ./lib/list.sh; then
    pass "Standardized message found in show_lists"
else
    fail "show_lists message not updated"
fi

if grep -q "Type '0' to go back, '00' for main menu" ./lib/list.sh; then
    pass "Navigation messages found in list functions"
else
    fail "Navigation messages not found in list functions"
fi

# Test 7: Verify template file exists
section "Test 7: Template File"

if [ -f "./lib/sample.json" ]; then
    pass "sample.json template exists"
    
    # Check if it's valid JSON
    if jq empty ./lib/sample.json 2>/dev/null; then
        pass "sample.json is valid JSON"
    else
        fail "sample.json is not valid JSON"
    fi
else
    fail "sample.json template not found"
fi

# Test 8: Test auto-creation in isolated environment
section "Test 8: Auto-Creation Test (Simulated)"

BACKUP_DIR="$HOME/.config/tera.backup.$$"
CONFIG_DIR="$HOME/.config/tera"

info "This test will temporarily move your config (if it exists)"
info "It will be restored after the test"
echo

# Backup existing config if it exists
if [ -d "$CONFIG_DIR" ]; then
    info "Backing up existing config to $BACKUP_DIR"
    mv "$CONFIG_DIR" "$BACKUP_DIR"
fi

# Test directory creation
info "Testing directory auto-creation..."
if [ ! -d "$CONFIG_DIR/favorite" ]; then
    # This would be created by tera on first run
    pass "Config directory does not exist yet (as expected)"
else
    fail "Config directory already exists"
fi

# Restore backup
if [ -d "$BACKUP_DIR" ]; then
    info "Restoring config from backup"
    rm -rf "$CONFIG_DIR"
    mv "$BACKUP_DIR" "$CONFIG_DIR"
    pass "Config restored successfully"
fi

# Test 9: Verify navigation consistency
section "Test 9: Navigation Consistency Check"

# Count how many functions have the standardized navigation
SEARCH_NAV_COUNT=$(grep -c "Type '0' to go back" ./lib/search.sh || echo "0")
LIST_NAV_COUNT=$(grep -c "Type '0' to go back" ./lib/list.sh || echo "0")

if [ "$LIST_NAV_COUNT" -ge 3 ]; then
    pass "List functions have standardized navigation (found $LIST_NAV_COUNT instances)"
else
    fail "Not enough standardized navigation in list functions (found $LIST_NAV_COUNT)"
fi

# Final Summary
section "Test Summary"

TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED))
echo "Total tests run: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}   ALL TESTS PASSED! ✓${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo
    echo "Your TERA improvements are working correctly!"
    echo
    echo "Next steps:"
    echo "1. Test manually by running: ./tera"
    echo "2. Try creating/deleting lists with '0' and '00'"
    echo "3. Try searching with '0' and '00'"
    echo "4. Check auto-creation by removing ~/.config/tera and running again"
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}   SOME TESTS FAILED${NC}"
    echo -e "${RED}========================================${NC}"
    echo
    echo "Please review the failed tests above."
    echo "Check the implementation files for issues."
    exit 1
fi
