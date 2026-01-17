#!/usr/bin/env bash
# Make executable: chmod +x test_station_improvements.sh

# Quick test script to verify station name improvements
# Run this after the updates to check functionality

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/lib.sh"

echo "================================"
echo "TERA Station Name Improvements"
echo "Testing Script"
echo "================================"
echo

# Test 1: Check if _station_list produces sorted output
echo "Test 1: Checking if stations are sorted alphabetically..."
echo "----------------------------------------"

# Check if there are any favorite lists
FAVORITE_PATH="${HOME}/.config/tera/favorite"
if [ -d "$FAVORITE_PATH" ]; then
    shopt -s nullglob
    json_files=("$FAVORITE_PATH"/*.json)
    
    if [ ${#json_files[@]} -eq 0 ]; then
        yellowprint "No JSON files found in $FAVORITE_PATH"
        echo
    else
        for json_file in "${json_files[@]}"; do
            if [ -f "$json_file" ] && jq -e 'length > 0' "$json_file" >/dev/null 2>&1; then
            basename="${json_file##*/}"
            list_name="${basename%.json}"
            echo "List: $list_name"
            
            # Get stations using the updated function
            stations=$(_station_list "$list_name")
            
            # Check if sorted
            sorted_stations=$(echo "$stations" | sort -f)
            
            if [ "$stations" = "$sorted_stations" ]; then
                greenprint "✓ Stations are properly sorted"
            else
                redprint "✗ Stations are NOT sorted"
                echo "Current order:"
                echo "$stations"
                echo ""
                echo "Expected order:"
                echo "$sorted_stations"
            fi
            echo
        fi
    done
    fi
else
    yellowprint "No favorite lists found at $FAVORITE_PATH"
fi

# Test 2: Check for whitespace in station names
echo "Test 2: Checking for whitespace in station names..."
echo "----------------------------------------"

if [ -d "$FAVORITE_PATH" ]; then
    whitespace_found=0
    for json_file in "$FAVORITE_PATH"/*.json; do
        if [ -f "$json_file" ]; then
            # Check for leading or trailing whitespace
            names_with_whitespace=$(jq -r '.[] | .name | select(test("^\\s|\\s$"))' "$json_file" 2>/dev/null)
            
            if [ -n "$names_with_whitespace" ]; then
                whitespace_found=1
                basename="${json_file##*/}"
                redprint "✗ Found whitespace in $basename:"
                echo "$names_with_whitespace"
                echo
            fi
        fi
    done
    
    if [ $whitespace_found -eq 0 ]; then
        greenprint "✓ No leading/trailing whitespace found in station names"
    fi
    echo
fi

# Test 3: Verify jq trimming pattern works
echo "Test 3: Testing jq trim pattern..."
echo "----------------------------------------"

test_string="  Test Station  "
trimmed=$(echo "$test_string" | jq -R 'gsub("^\\s+|\\s+$";"")')
expected='"Test Station"'

if [ "$trimmed" = "$expected" ]; then
    greenprint "✓ Trimming pattern works correctly"
else
    redprint "✗ Trimming pattern failed"
    echo "Input: '$test_string'"
    echo "Output: $trimmed"
    echo "Expected: $expected"
fi
echo

# Test 4: Check alphabetical sorting (case-insensitive)
echo "Test 4: Testing case-insensitive alphabetical sort..."
echo "----------------------------------------"

test_input="Zebra
apple
Banana
cherry"

sorted_output=$(echo "$test_input" | sort -f)
expected_output="apple
Banana
cherry
Zebra"

if [ "$sorted_output" = "$expected_output" ]; then
    greenprint "✓ Case-insensitive sorting works correctly"
else
    redprint "✗ Case-insensitive sorting failed"
    echo "Input:"
    echo "$test_input"
    echo ""
    echo "Output:"
    echo "$sorted_output"
    echo ""
    echo "Expected:"
    echo "$expected_output"
fi
echo

echo "================================"
echo "Testing Complete!"
echo "================================"
