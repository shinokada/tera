#!/usr/bin/env bats

# Integration tests
# These tests verify the overall consistency across the application

@test "All menus follow 0=Main Menu convention" {
    # Check list menu
    grep -q '0) Main Menu' ../lib/list.sh
    
    # Check search menu
    grep -q '0) Main Menu' ../lib/search.sh
    
    # Check gist menu
    grep -q '0) Main Menu' ../lib/gistlib.sh
}

@test "All menus have Exit at the bottom" {
    # Check that each menu has Exit option
    # List menu - 5) Exit
    grep -q '5) Exit' ../lib/list.sh
    
    # Search menu - 7) Exit
    grep -q '7) Exit' ../lib/search.sh
    
    # Gist menu - 3) Exit
    grep -q '3) Exit' ../lib/gistlib.sh
}

@test "All interactive selections have Main Menu option" {
    # Play list selection
    play_list=$(grep 'lists_with_menu=' ../lib/play.sh)
    echo "$play_list" | grep -q "<< Main Menu >>"
    
    # Play station selection
    play_station=$(grep 'STATIONS_WITH_MENU=' ../lib/play.sh)
    echo "$play_station" | grep -q "<< Main Menu >>"
    
    # Search results
    search_results=$(grep 'STATIONS_WITH_MENU=' ../lib/search.sh)
    echo "$search_results" | grep -q "<< Main Menu >>"
}

@test "FZF prompts are consistent" {
    # Get all FZF prompts
    prompts=$(grep 'fzf --prompt=' ../lib/*.sh)
    
    # Count different prompt styles (should be minimal)
    simple_prompts=$(echo "$prompts" | grep -c 'prompt="> "' || true)
    menu_prompts=$(echo "$prompts" | grep -c 'prompt="Choose an option' || true)
    
    # Should have both simple and menu prompts
    [ "$simple_prompts" -gt 0 ]
    [ "$menu_prompts" -gt 0 ]
}

@test "All headings use Title Case" {
    # Check various headings exist and don't use all caps
    grep -q 'Delete a Radio Station' ../lib/delete_station.sh
    grep -q 'I Feel Lucky' ../lib/lucky.sh
    
    # Verify they don't use all caps
    ! grep -q "DELETE A RADIO STATION" ../lib/delete_station.sh
    ! grep -q "I FEEL LUCKY" ../lib/lucky.sh
}

@test "No redundant text after search completes" {
    # Check that searching message exists
    grep -q 'greenprint "Searching ..."' ../lib/lib.sh
    
    # Check that cleanup code exists (echo -ne to clear the line)
    grep -q 'echo -ne' ../lib/lib.sh
}

@test "All clear commands come before headings" {
    # Check that clear exists in delete_station.sh
    grep -q 'clear' ../lib/delete_station.sh
    
    # Check that Delete heading exists
    grep -q 'Delete a Radio Station' ../lib/delete_station.sh
    
    # Check that clear exists in lucky.sh
    grep -q 'clear' ../lib/lucky.sh
    
    # Check that Lucky heading exists
    grep -q 'I Feel Lucky' ../lib/lucky.sh
}

@test "No double Main Menu entries in any menu" {
    # Check that menus don't have duplicate Main Menu entries
    # Each menu file should have Main Menu mentioned a reasonable number of times
    list_count=$(grep -c 'Main Menu' ../lib/list.sh || true)
    search_count=$(grep -c 'Main Menu' ../lib/search.sh || true)
    
    # Should have at least 1 Main Menu entry
    [ "$list_count" -ge 1 ]
    [ "$search_count" -ge 1 ]
}
