#!/usr/bin/env bats

# Test heading displays
# These tests verify that all pages have appropriate headings

setup() {
    export APP_NAME="TERA"
}

@test "Play from my list has heading" {
    # Check that header is defined in play.sh
    grep -q 'header=' ../lib/play.sh
    
    # Check that it contains the expected text
    grep -q "Play from My List" ../lib/play.sh
}

@test "Delete station has heading" {
    # Check that Delete heading exists
    grep -q 'Delete a Radio Station' ../lib/delete_station.sh
    
    # Check that it uses cyanprint
    result=$(grep 'Delete a Radio Station' ../lib/delete_station.sh)
    echo "$result" | grep -q "cyanprint"
}

@test "I Feel Lucky has heading" {
    # Check that Lucky heading exists
    grep -q 'I Feel Lucky' ../lib/lucky.sh
    
    # Check that it uses cyanprint
    result=$(grep 'I Feel Lucky' ../lib/lucky.sh)
    echo "$result" | grep -q "cyanprint"
}

@test "Search by functions have headings" {
    # Check that Search by text exists
    grep -q 'Search by' ../lib/search.sh
}

@test "Advanced search has heading" {
    # Check that Advanced Search heading exists
    grep -q 'Advanced Search' ../lib/search.sh
    
    # Check that it uses cyanprint
    result=$(grep 'Advanced Search' ../lib/search.sh)
    echo "$result" | grep -q "cyanprint"
}

@test "Create gist has heading" {
    # Check that Create gist heading exists
    grep -q 'Create a Gist' ../lib/gistlib.sh
    
    # Check that it uses cyanprint
    result=$(grep 'Create a Gist' ../lib/gistlib.sh)
    echo "$result" | grep -q "cyanprint"
}

@test "Recover gist has heading" {
    # Check that Recover heading exists
    grep -q 'Recover Favorites' ../lib/gistlib.sh
    
    # Check that it uses cyanprint
    result=$(grep 'Recover Favorites' ../lib/gistlib.sh)
    echo "$result" | grep -q "cyanprint"
}

@test "All headings use cyanprint function" {
    # Check that headings consistently use cyanprint
    delete_heading=$(grep 'Delete a Radio Station' ../lib/delete_station.sh)
    lucky_heading=$(grep 'I Feel Lucky' ../lib/lucky.sh)
    
    echo "$delete_heading" | grep -q "cyanprint"
    echo "$lucky_heading" | grep -q "cyanprint"
}

@test "FZF headers use header-first flag" {
    # Check that header-first flag is used in play.sh
    grep -q 'header-first' ../lib/play.sh
}
