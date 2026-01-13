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
    # Extract menu options from main tera script
    result=$(grep -A 10 'MENU_OPTIONS=' ../tera | head -15)
    
    # Check that it contains the expected options
    echo "$result" | grep -q "1) Play from my list"
    echo "$result" | grep -q "0) Exit"
}

@test "List menu has Main Menu at position 0" {
    result=$(grep -A 10 'list_menu()' ../lib/list.sh | grep -A 10 'MENU_OPTIONS=')
    
    # Check that Main Menu is at position 0
    echo "$result" | grep -q "0) Main Menu"
    echo "$result" | grep -q "1) Create a list"
    echo "$result" | grep -q "5) Exit"
}

@test "Search menu has Main Menu at position 0" {
    result=$(grep -A 15 'search_menu()' ../lib/search.sh | grep -A 15 'MENU_OPTIONS=')
    
    # Check that Main Menu is at position 0
    echo "$result" | grep -q "0) Main Menu"
    echo "$result" | grep -q "1) Tag"
    echo "$result" | grep -q "7) Exit"
}

@test "Search submenu has Main Menu at position 0" {
    result=$(grep -A 10 'search_submenu()' ../lib/search.sh | grep -A 10 'MENU_OPTIONS=')
    
    # Check that Main Menu is at position 0
    echo "$result" | grep -q "0) Main Menu"
    echo "$result" | grep -q "1) Play"
    echo "$result" | grep -q "4) Exit"
}

@test "Gist menu has Main Menu at position 0" {
    result=$(grep -A 10 'gist_menu()' ../lib/gistlib.sh | grep -A 10 'MENU_OPTIONS=')
    
    # Check that Main Menu is at position 0
    echo "$result" | grep -q "0) Main Menu"
    echo "$result" | grep -q "1) Create a gist"
    echo "$result" | grep -q "3) Exit"
}

@test "Play function has Main Menu option in list selection" {
    result=$(grep -A 5 'lists_with_menu=' ../lib/play.sh)
    
    # Check that Main Menu option is added
    echo "$result" | grep -q "<< Main Menu >>"
}

@test "Play function has Main Menu option in station selection" {
    result=$(grep -A 5 'STATIONS_WITH_MENU=' ../lib/play.sh)
    
    # Check that Main Menu option is added
    echo "$result" | grep -q "<< Main Menu >>"
}

@test "Search results have Main Menu option" {
    result=$(grep -A 5 'STATIONS_WITH_MENU=' ../lib/search.sh)
    
    # Check that Main Menu option is added in search results
    echo "$result" | grep -q "<< Main Menu >>"
}

@test "Delete station has Main Menu option" {
    result=$(grep 'Main Menu' ../lib/delete_station.sh)
    
    # Check that Main Menu text exists
    echo "$result" | grep -q "Main Menu"
}

@test "All menus use consistent prompt style" {
    # Check that prompts use "> " or are consistent
    play_prompt=$(grep 'fzf --prompt=' ../lib/play.sh | head -1)
    search_prompt=$(grep 'fzf --prompt=' ../lib/search.sh | head -1)
    
    # Both should use simple prompts
    echo "$play_prompt" | grep -q 'prompt="> "'
    echo "$search_prompt" | grep -q 'prompt="> "'
}
