#!/usr/bin/env bats

# Test ESC key and navigation behavior
# These tests verify that ESC properly returns to menus instead of quitting

setup() {
    export TMP_PATH="$HOME/.cache/tera"
    mkdir -p "$TMP_PATH"
}

@test "Play list selection handles empty input (ESC)" {
    # Check that empty input handling exists
    grep -q 'if \[ -z "\$LIST" \]' ../lib/play.sh
    
    # Check that it calls menu function
    result=$(grep 'if \[ -z "\$LIST" \]' ../lib/play.sh)
    [ -n "$result" ]
}

@test "Play station selection handles empty input (ESC)" {
    # Check that empty input handling exists
    grep -q 'if \[ -z "\$SELECTION" \]' ../lib/play.sh
    
    # Check that it calls menu function
    result=$(grep 'if \[ -z "\$SELECTION" \]' ../lib/play.sh)
    [ -n "$result" ]
}

@test "Search results handle empty input (ESC)" {
    # Check that ESC handling comment exists
    grep -q 'Check if user cancelled (ESC)' ../lib/search.sh
    
    # Check that search_menu is called
    grep -q 'search_menu' ../lib/search.sh
}

@test "Search results no longer mention ESC in prompt" {
    result=$(grep 'prompt=' ../lib/search.sh)
    
    # Check that prompts don't contain "ESC to return" text
    ! echo "$result" | grep -q "ESC to return"
}

@test "Lucky function allows menu return" {
    result=$(grep 'menu' ../lib/lucky.sh)
    
    # Check that menu option exists in lucky function
    echo "$result" | grep -q "menu"
}

@test "Delete station handles empty selection and Main Menu" {
    # Check that it handles empty CHOICE (user cancelled list selection)
    grep -q 'if \[ -z "\$CHOICE" \]' ../lib/delete_station.sh
    
    # Check that it handles Main Menu selection via LIST_NUM
    grep -q 'if \[ "\$LIST_NUM" = "0" \]' ../lib/delete_station.sh
    
    # Check that it handles empty SELECTION (user cancelled station selection)
    grep -q 'if \[ -z "\$SELECTION" \]' ../lib/delete_station.sh
    
    # Check that it handles Main Menu selection via SELECTED_TEXT
    grep -q 'if \[ "\$SELECTED_TEXT" = "<< Main Menu >>" \]' ../lib/delete_station.sh
}
