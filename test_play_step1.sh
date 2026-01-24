#!/bin/bash

echo "Testing Play Screen - Step 1: List Selection"
echo "=============================================="
echo

echo "Running tests..."
cd /Users/shinichiokada/Terminal-Tools/tera
go test ./internal/ui/... -v -run "TestPlayModel|TestGetAvailableLists"
