#!/bin/bash
# Emergency bug fix test script

echo "════════════════════════════════════════════════════════"
echo "  TERA Critical Bug Fixes - Test Script"
echo "════════════════════════════════════════════════════════"
echo ""

# 1. Kill any existing mpv processes
echo "1. Cleaning up any existing mpv processes..."
killall mpv 2>/dev/null
sleep 1

count=$(ps aux | grep mpv | grep -v grep | wc -l)
if [ $count -eq 0 ]; then
    echo "   ✓ No mpv processes running"
else
    echo "   ✗ Warning: $count mpv processes still running"
    echo "   Trying force kill..."
    pkill -9 mpv
    sleep 1
fi
echo ""

# 2. Build
echo "2. Building application..."
if go build -o tera ./cmd/tera/; then
    echo "   ✓ Build successful"
else
    echo "   ✗ Build failed"
    exit 1
fi
echo ""

# 3. Check binary
echo "3. Verifying binary..."
if [ -f "./tera" ]; then
    echo "   ✓ Binary exists"
else
    echo "   ✗ Binary not found"
    exit 1
fi
echo ""

# 4. Test instructions
echo "════════════════════════════════════════════════════════"
echo "  Ready for Testing!"
echo "════════════════════════════════════════════════════════"
echo ""
echo "CRITICAL TESTS:"
echo ""
echo "Test 1: Single Station"
echo "  1. Run: ./tera"
echo "  2. Search and play a station"
echo "  3. In another terminal: ps aux | grep mpv"
echo "  4. Verify: Only 1 mpv process"
echo ""
echo "Test 2: Switch Stations"
echo "  1. Play station A"
echo "  2. Press Esc to go back"
echo "  3. Play station B"
echo "  4. Check: ps aux | grep mpv"
echo "  5. Verify: Still only 1 mpv process"
echo "  6. Verify: Only station B is heard"
echo ""
echo "Test 3: Navigation Cleanup"
echo "  1. Play a station"
echo "  2. Press Esc to go back"
echo "  3. Check: ps aux | grep mpv"
echo "  4. Verify: No mpv process (stopped)"
echo ""
echo "Test 4: Station Info"
echo "  1. Search stations"
echo "  2. Select one with Enter"
echo "  3. Verify: Shows ONLY selected station"
echo "  4. Test arrow keys in menu"
echo "  5. Verify: Menu navigation works"
echo ""
echo "Test 5: No Panic"
echo "  1. Play, stop, play different stations"
echo "  2. Navigate back and forth"
echo "  3. Verify: No panic errors"
echo "  4. Verify: No SIGSEGV errors"
echo ""
echo "════════════════════════════════════════════════════════"
echo ""
echo "MONITORING COMMAND (run in another terminal):"
echo ""
echo "  watch 'ps aux | grep mpv | grep -v grep | wc -l'"
echo ""
echo "  Should show: 0 (not playing) or 1 (playing)"
echo "  NEVER MORE THAN 1!"
echo ""
echo "════════════════════════════════════════════════════════"
echo ""
echo "Run ./tera to start testing"
echo ""
