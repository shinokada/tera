#!/bin/bash
# Quick build and basic verification

set -e

echo "🔧 Building tera with bug fixes..."
make clean
make build

if [ ! -f "./tera" ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"
echo ""
echo "📋 Fixed Issues:"
echo "  1. ✅ Station stops when quitting"
echo "  2. ✅ Search menu shows all options (height fixed)"
echo "  3. ✅ Save prompt after search playback"
echo "  4. ✅ Filter count shows 'x/y items'"
echo "  5. ✅ Play screen uses full height"
echo ""
echo "🧪 Quick Tests:"
echo ""

# Test 1: Binary exists and is executable
if [ -x "./tera" ]; then
    echo "✅ Binary is executable"
else
    echo "❌ Binary is not executable"
    exit 1
fi

# Test 2: Check for required dependencies
if command -v mpv &> /dev/null; then
    echo "✅ MPV is installed"
else
    echo "⚠️  Warning: MPV not found (required for playback)"
fi

echo ""
echo "📖 Testing Instructions:"
echo ""
echo "Test 1 - Quit Stops Player:"
echo "  • Run: ./tera"
echo "  • Navigate to Search (2) or Play (1)"
echo "  • Play a station"
echo "  • Press 'q' to quit"
echo "  • Verify: Audio stops immediately"
echo "  • Check: ps aux | grep mpv (should be empty)"
echo ""
echo "Test 2 - Screen Heights:"
echo "  • Run: ./tera"
echo "  • Press 2 (Search)"
echo "  • Verify: All 6 search options visible"
echo "  • Go back, press 1 (Play)"
echo "  • Verify: List uses most of screen"
echo ""
echo "Test 3 - Save Prompt:"
echo "  • Search for stations"
echo "  • Select and play one"
echo "  • Press 'q'"
echo "  • Verify: Save prompt appears"
echo "  • Try both 'y' and 'n' options"
echo ""
echo "Test 4 - Filter Count:"
echo "  • Search for stations (get results)"
echo "  • Press '/' to filter"
echo "  • Type some text"
echo "  • Verify: Status bar shows 'x/y items'"
echo ""
echo "Test 5 - Window Resize:"
echo "  • Run: ./tera"
echo "  • Resize terminal window"
echo "  • Navigate through screens"
echo "  • Verify: Lists adapt to new size"
echo ""
echo "🚀 Ready to test! Run: ./tera"
echo ""
echo "📝 Report any issues found during testing"
