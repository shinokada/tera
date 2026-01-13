#!/usr/bin/env bats

# Test ESC key and navigation behavior
# These tests verify that ESC properly returns to menus instead of quitting

setup() {
    export TMP_PATH="$HOME/.cache/tera"
    mkdir -p "$TMP_PATH"
}

@test "Play list selection handles empty input (ESC)" {
    result=$(grep -A 3 'if \[ -z "\$LIST" \]' ../lib/play.sh | head -5)
    
    # Check that empty input calls menu function
    echo "$result" | grep -q "menu"
}

@test "Play station selection handles empty input (ESC)" {
    result=$(grep -A 3 'if \[ -z "\$SELECTION" \]' ../lib/play.sh | head -5)
    
    # Check that empty input calls menu function
    echo "$result" | grep -q "menu"
}

@test "Search results handle empty input (ESC)" {
    result=$(grep -A 3 'Check if user cancelled (ESC)' ../lib/search.sh | head -5)
    
    # Check that ESC returns to search menu
    echo "$result" | grep -q "search_menu"
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

@test "Delete station handles zero input for Main Menu" {
    result=$(grep 'if \[\[ -z \$ANS \]\] || \[\[ \$ANS == "0" \]\]' ../lib/delete_station.sh)
    
    # Check that condition exists
    [ -n "$result" ]
}
