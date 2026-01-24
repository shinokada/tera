#!/usr/bin/env bats

# Test duplicate detection when saving stations
# Verifies that the same station cannot be saved twice to the same list

setup() {
    export SCRIPT_DOT_DIR="$HOME/.config/tera"
    export FAVORITE_PATH="$SCRIPT_DOT_DIR/favorite"
    export TMP_PATH="$HOME/.cache/tera"
    export APP_NAME="TERA"
    
    # Create test directory structure
    mkdir -p "$FAVORITE_PATH"
    mkdir -p "$TMP_PATH"
}

teardown() {
    # Clean up test files if needed
    rm -f "$TMP_PATH"/test_*
}

@test "Save station function checks for duplicate stationuuid" {
    # Check that the function extracts stationuuid for comparison
    grep -q 'STATION_UUID.*stationuuid' ../lib/search.sh
}

@test "Save station function queries existing stations by uuid" {
    # Check that the function searches for existing station by uuid
    result=$(grep 'EXISTING_STATION.*stationuuid.*uuid' ../lib/search.sh)
    [ -n "$result" ]
}

@test "Save station function shows warning for duplicates" {
    # Check that a warning message is shown when duplicate is found
    grep -q 'already in your.*list' ../lib/search.sh
}

@test "Save station function returns early on duplicate" {
    # Verify that the function returns without saving when duplicate is found
    # Check that there's a return statement after the duplicate check
    awk '/if.*EXISTING_STATION/,/fi/' ../lib/search.sh | grep -q 'return'
}

@test "Duplicate check uses yellowprint for warning" {
    # Check that the duplicate warning uses yellowprint (consistent with app style)
    awk '/if.*EXISTING_STATION/,/fi/' ../lib/search.sh | grep -q 'yellowprint'
}

@test "Duplicate check happens before jq add operation" {
    # Verify the order: extract station, check duplicate, then add
    # Get line numbers
    uuid_line=$(grep -n 'STATION_UUID.*stationuuid' ../lib/search.sh | head -n1 | cut -d: -f1)
    add_line=$(grep -n 'jq.*+= \[input\]' ../lib/search.sh | head -n1 | cut -d: -f1)
    
    # UUID extraction should come before the add operation
    [ "$uuid_line" -lt "$add_line" ]
}
