#!/bin/bash

echo "=== Step 1: Find the success message and next 10 lines ==="
grep -A10 'Successfully created a secret Gist' ../lib/gistlib.sh
echo ""
echo "=== Step 2: From above, find gist_menu and next line ==="
grep -A10 'Successfully created a secret Gist' ../lib/gistlib.sh | grep -A1 'gist_menu'
echo ""
echo "=== Step 3: Check if return is in there ==="
grep -A10 'Successfully created a secret Gist' ../lib/gistlib.sh | grep -A1 'gist_menu' | grep 'return'
if [ $? -eq 0 ]; then
    echo "✓ FOUND return"
else
    echo "✗ NOT FOUND return"
fi

echo ""
echo "=== Let's try with more context (20 lines) ==="
grep -A20 'Successfully created a secret Gist' ../lib/gistlib.sh | grep -A1 'gist_menu' | grep 'return'
if [ $? -eq 0 ]; then
    echo "✓ FOUND return with -A20"
else
    echo "✗ NOT FOUND return with -A20"
fi
