#!/usr/bin/env bats

# Test menu structure and navigation
# These tests verify that all menus have consistent structure with Main Menu at position 0

setup() {
    # Source the lib files
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

@test "Main menu has correct structure" {
    # Check that main menu contains expected options
    grep -q '1) Play from my list' ../tera
    grep -q '0) Exit' ../tera
}

@test "List menu has Main Menu at position 0" {
    # Check for list_menu function
    grep -q 'list_menu()' ../lib/list.sh
    
    # Check that Main Menu is at position 0
    grep -q '0) Main Menu' ../lib/list.sh
    grep -q '1) Create a list' ../lib/list.sh
    grep -q '5) Exit' ../lib/list.sh
}

@test "Search menu has Main Menu at position 0" {
    # Check for search_menu function
    grep -q 'search_menu()' ../lib/search.sh
    
    # Check that Main Menu is at position 0
    grep -q '0) Main Menu' ../lib/search.sh
    grep -q '1) Tag' ../lib/search.sh
    grep -q '7) Exit' ../lib/search.sh
}

@test "Search submenu has Main Menu at position 0" {
    # Check for search_submenu function
    grep -q 'search_submenu()' ../lib/search.sh
    
    # Check that Main Menu is at position 0 in submenu
    awk '/search_submenu\(\)/,/^[[:space:]]*}/' ../lib/search.sh | grep -q '0) Main Menu'
}

@test "Gist menu has Main Menu at position 0" {
    # Check for gist_menu function
    grep -q 'gist_menu()' ../lib/gistlib.sh
    
    # Check that Main Menu is at position 0
    grep -q '0) Main Menu' ../lib/gistlib.sh
    grep -q '1) Create a gist' ../lib/gistlib.sh
    grep -q '6) Exit' ../lib/gistlib.sh
}

@test "Play function has Main Menu option in list selection" {
    result=$(grep 'lists_with_menu=' ../lib/play.sh)
    
    # Check that Main Menu option is added
    echo "$result" | grep -q "<< Main Menu >>"
}

@test "Play function has Main Menu option in station selection" {
    result=$(grep 'STATIONS_WITH_MENU=' ../lib/play.sh)
    
    # Check that Main Menu option is added
    echo "$result" | grep -q "<< Main Menu >>"
}

@test "Search results have Main Menu option" {
    result=$(grep 'STATIONS_WITH_MENU=' ../lib/search.sh)
    
    # Check that Main Menu option is added in search results
    echo "$result" | grep -q "<< Main Menu >>"
}

@test "Delete station has Main Menu option" {
    result=$(grep 'Main Menu' ../lib/delete_station.sh)
    
    # Check that Main Menu text exists
    echo "$result" | grep -q "Main Menu"
}

@test "All menus use fzf for interactive selection" {
    # Check that play.sh uses fzf with prompts
    play_result=$(grep 'fzf --prompt=' ../lib/play.sh)
    [ -n "$play_result" ]
    
    # Check that search.sh uses fzf with prompts
    search_result=$(grep 'fzf --prompt=' ../lib/search.sh)
    [ -n "$search_result" ]
    
    # Verify that at least some functions use the simple "> " prompt
    grep -q 'fzf --prompt="> "' ../lib/search.sh
    grep -q 'fzf --prompt="> "' ../lib/play.sh
    
    # Verify delete_station.sh and gistlib.sh also use fzf
    grep -q 'fzf --prompt=' ../lib/delete_station.sh
    grep -q 'fzf --prompt=' ../lib/gistlib.sh
}
