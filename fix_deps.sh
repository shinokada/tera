#!/bin/bash

echo "Fixing Go dependencies..."
cd /Users/shinichiokada/Terminal-Tools/tera

# Clean the module cache
echo "Cleaning module cache..."
go clean -modcache

# Tidy up dependencies
echo "Running go mod tidy..."
go mod tidy

# Download all dependencies
echo "Downloading dependencies..."
go mod download

# Get the missing packages explicitly
echo "Getting missing packages..."
go get github.com/atotto/clipboard
go get github.com/sahilm/fuzzy

# Tidy again
echo "Running go mod tidy again..."
go mod tidy

echo "Done! Now try running the tests."
