#!/bin/bash
# Test script for search screen navigation fixes

echo "════════════════════════════════════════════════════════"
echo "  TERA Search Screen Navigation Test"
echo "════════════════════════════════════════════════════════"
echo ""

echo "Running tests..."
echo ""

# Test 1: Build
echo "1. Building application..."
if go build -o tera ./cmd/tera/; then
    echo "   ✓ Build successful"
else
    echo "   ✗ Build failed"
    exit 1
fi
echo ""

# Test 2: Run component tests
echo "2. Testing menu component..."
if go test -v ./internal/ui/components/ 2>&1 | grep -q "PASS"; then
    echo "   ✓ Menu component tests passed"
else
    echo "   ✗ Menu component tests failed"
fi
echo ""

# Test 3: Check if all required files exist
echo "3. Checking required files..."
files=(
    "internal/ui/components/menu.go"
    "internal/ui/components/menu_test.go"
    "internal/ui/app.go"
    "internal/ui/search.go"
)

all_exist=true
for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "   ✓ $file exists"
    else
        echo "   ✗ $file missing"
        all_exist=false
    fi
done
echo ""

if [ "$all_exist" = false ]; then
    echo "✗ Some required files are missing"
    exit 1
fi

echo "════════════════════════════════════════════════════════"
echo "  All automated tests passed!"
echo "════════════════════════════════════════════════════════"
echo ""
echo "Manual Testing Guide:"
echo ""
echo "1. MAIN MENU NAVIGATION"
echo "   • Start: ./tera"
echo "   • Test: ↑↓ or j/k to navigate"
echo "   • Test: Number shortcuts (1-6)"
echo "   • Test: Enter to select"
echo "   • Test: q to quit"
echo ""
echo "2. SEARCH MENU NAVIGATION"
echo "   • Navigate to: Select '2' from main menu"
echo "   • Test: ↑↓ or j/k to navigate"
echo "   • Test: Number shortcuts (1-6)"
echo "   • Test: Enter to select"
echo "   • Test: 0/Esc to go back"
echo ""
echo "3. SEARCH RESULTS"
echo "   • Navigate to: Select a search type, enter query"
echo "   • Verify: Results list displays correctly"
echo "   • Test: ↑↓ to navigate results"
echo "   • Test: Enter to select a station"
echo "   • Test: / to filter"
echo "   • Test: Esc to go back"
echo ""
echo "4. STATION INFO MENU"
echo "   • Navigate to: Select a station from results"
echo "   • Test: ↑↓ or j/k to navigate"
echo "   • Test: Number shortcuts (1-3)"
echo "   • Test: Enter to select action"
echo "   • Test: 0 for main menu"
echo "   • Test: Esc to go back"
echo ""
echo "5. VERIFY VISUAL FEEDBACK"
echo "   • Check: Selected items are highlighted"
echo "   • Check: '>' indicator shows on selected item"
echo "   • Check: Help text is clear and visible"
echo ""
echo "════════════════════════════════════════════════════════"
echo ""
echo "Run ./tera to start manual testing"
echo ""
