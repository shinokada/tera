#!/bin/bash

cd /Users/shinichiokada/Terminal-Tools/tera

echo "================================"
echo "Running Play Screen Step 3 Tests"
echo "================================"
echo ""

echo "Testing MPV Player..."
go test ./internal/player -v

echo ""
echo "Testing Play Screen..."
go test ./internal/ui -v -run Play

# Check result
if [ $? -eq 0 ]; then
    echo ""
    echo "================================"
    echo "✅ All tests passed!"
    echo "================================"
    echo ""
    echo "Next steps:"
    echo "1. Make sure MPV is installed: brew install mpv"
    echo "2. Build: go build -o tera cmd/tera/main.go"
    echo "3. Create test station data (see PLAY_SCREEN_STEP3.md)"
    echo "4. Run: ./tera"
    echo "5. Play a real stream!"
else
    echo ""
    echo "================================"
    echo "❌ Tests failed"
    echo "================================"
fi
