#!/bin/bash

cd /Users/shinichiokada/Terminal-Tools/tera

echo "================================"
echo "Running Play Screen Step 2 Tests"
echo "================================"
echo ""

# Run all Play tests
go test ./internal/ui -v -run "Play|Station"

# Check result
if [ $? -eq 0 ]; then
    echo ""
    echo "================================"
    echo "✅ All tests passed!"
    echo "================================"
    echo ""
    echo "Next steps:"
    echo "1. Build: go build -o tera cmd/tera/main.go"
    echo "2. Create test data with stations:"
    echo "   See test_data_example.json below"
    echo "3. Run: ./tera"
    echo "4. Press '1' to enter Play screen"
    echo "5. Select a list, then filter stations with '/'"
else
    echo ""
    echo "================================"
    echo "❌ Tests failed"
    echo "================================"
fi
