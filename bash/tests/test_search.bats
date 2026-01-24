#!/usr/bin/env bats

# Test search functionality improvements
# These tests verify search message cleanup and menu options

setup() {
    export TMP_PATH="$HOME/.cache/tera"
    mkdir -p "$TMP_PATH"
}

@test "wget_simple_search clears 'Searching...' message" {
    result=$(grep '_wget_simple_search()' ../lib/lib.sh)
    
    # Check that function exists
    [ -n "$result" ]
    
    # Check that cleanup code exists in the function
    awk '/_wget_simple_search\(\)/,/^}/' ../lib/lib.sh | grep -q 'echo -ne'
}

@test "wget_search clears 'Searching...' message" {
    result=$(grep '_wget_search()' ../lib/lib.sh)
    
    # Check that function exists
    [ -n "$result" ]
    
    # Check that cleanup code exists in the function
    awk '/_wget_search\(\)/,/^}/' ../lib/lib.sh | grep -q 'echo -ne'
}

@test "Search results include Main Menu option" {
    result=$(grep 'STATIONS_WITH_MENU=' ../lib/search.sh)
    
    # Check that Main Menu is added to search results
    echo "$result" | grep -q "<< Main Menu >>"
}

@test "Search results adjust station numbers correctly" {
    result=$(grep 'ANS=\$((ANS - 1))' ../lib/search.sh)
    
    # Check that station numbers are adjusted for Main Menu offset
    [ -n "$result" ]
}

@test "Advanced search includes Main Menu option" {
    # Count occurrences of STATIONS_WITH_MENU in search.sh
    count=$(grep -c 'STATIONS_WITH_MENU=' ../lib/search.sh)
    
    # Should have at least 2 (search_by and advanced_search)
    [ "$count" -ge 2 ]
}

@test "Search functions use consistent headers" {
    result=$(grep 'header=".*Search' ../lib/search.sh)
    
    # Check that search functions use headers
    [ -n "$result" ]
}
