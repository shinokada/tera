#!/usr/bin/env bats

# Test heading displays
# These tests verify that all pages have appropriate headings

setup() {
    export APP_NAME="TERA"
}

@test "Play from my list has heading" {
    result=$(grep 'header=' ../lib/play.sh | head -1)
    
    # Check that header contains the app name and feature name
    echo "$result" | grep -q "Play from My List"
}

@test "Delete station has heading" {
    result=$(grep 'cyanprint.*Delete a Radio Station' ../lib/delete_station.sh)
    
    # Check that heading exists
    [ -n "$result" ]
}

@test "I Feel Lucky has heading" {
    result=$(grep 'cyanprint.*I Feel Lucky' ../lib/lucky.sh)
    
    # Check that heading exists
    [ -n "$result" ]
}

@test "Search by functions have headings" {
    result=$(grep 'Search by' ../lib/search.sh)
    
    # Check that search headings exist
    echo "$result" | grep -q "Search by"
}

@test "Advanced search has heading" {
    result=$(grep 'cyanprint.*Advanced Search' ../lib/search.sh)
    
    # Check that heading exists
    [ -n "$result" ]
}

@test "Create gist has heading" {
    result=$(grep 'cyanprint.*Create a Gist' ../lib/gistlib.sh)
    
    # Check that heading exists
    [ -n "$result" ]
}

@test "Recover gist has heading" {
    result=$(grep 'cyanprint.*Recover Favorites' ../lib/gistlib.sh)
    
    # Check that heading exists
    [ -n "$result" ]
}

@test "All headings use cyanprint function" {
    # Check that headings consistently use cyanprint
    delete_heading=$(grep 'Delete a Radio Station' ../lib/delete_station.sh)
    lucky_heading=$(grep 'I Feel Lucky' ../lib/lucky.sh)
    
    echo "$delete_heading" | grep -q "cyanprint"
    echo "$lucky_heading" | grep -q "cyanprint"
}

@test "FZF headers use header-first flag" {
    result=$(grep 'header-first' ../lib/play.sh)
    
    # Check that header-first flag is used
    [ -n "$result" ]
}
