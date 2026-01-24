#!/bin/bash

cd /Users/shinichiokada/Terminal-Tools/tera

echo "Adding missing dependencies..."

# Get the specific missing packages
go get github.com/atotto/clipboard
go get github.com/sahilm/fuzzy

# Tidy up
go mod tidy

echo "Dependencies fixed! Now running tests..."
echo ""

go test ./internal/ui -v -run Play
