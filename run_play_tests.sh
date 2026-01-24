#!/bin/bash

cd /Users/shinichiokada/Terminal-Tools/tera

echo "================================"
echo "Running Play Screen Tests"
echo "================================"
echo ""

# Run the tests
go test ./internal/ui -v -run Play

# Check result
if [ $? -eq 0 ]; then
    echo ""
    echo "================================"
    echo "✅ All tests passed!"
    echo "================================"
    echo ""
    echo "Next steps:"
    echo "1. Try building: go build -o tera cmd/tera/main.go"
    echo "2. Run the app: ./tera"
    echo "3. Press '1' to enter Play screen"
else
    echo ""
    echo "================================"
    echo "❌ Tests failed"
    echo "================================"
fi
