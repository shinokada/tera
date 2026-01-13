#!/usr/bin/env bats

# Integration tests
# These tests verify the overall consistency across the application

@test "All menus follow 0=Main Menu convention" {
    # Check list menu
    list_menu=$(grep -A 10 'list_menu()' ../lib/list.sh | grep 'MENU_OPTIONS=')
    echo "$list_menu" | grep -q "0) Main Menu"
    
    # Check search menu
    search_menu=$(grep -A 10 'search_menu()' ../lib/search.sh | grep 'MENU_OPTIONS=')
    echo "$search_menu" | grep -q "0) Main Menu"
    
    # Check gist menu
    gist_menu=$(grep -A 10 'gist_menu()' ../lib/gistlib.sh | grep 'MENU_OPTIONS=')
    echo "$gist_menu" | grep -q "0) Main Menu"
    
    # Check search submenu
    search_submenu=$(grep -A 10 'search_submenu()' ../lib/search.sh | grep 'MENU_OPTIONS=')
    echo "$search_submenu" | grep -q "0) Main Menu"
}

@test "All menus have Exit at the bottom" {
    # Check that each menu's MENU_OPTIONS has Exit at the bottom
    # List menu - 5) Exit
    list_menu=$(grep -A 15 'list_menu()' ../lib/list.sh | grep -A 15 'MENU_OPTIONS=')
    echo "$list_menu" | grep -q "5) Exit"
    
    # Search menu - 7) Exit
    search_menu=$(grep -A 15 'search_menu()' ../lib/search.sh | grep -A 15 'MENU_OPTIONS=')
    echo "$search_menu" | grep -q "7) Exit"
    
    # Gist menu - 3) Exit
    gist_menu=$(grep -A 15 'gist_menu()' ../lib/gistlib.sh | grep -A 15 'MENU_OPTIONS=')
    echo "$gist_menu" | grep -q "3) Exit"
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
    prompts=$(grep -h 'fzf --prompt=' ../lib/*.sh)
    
    # Count different prompt styles (should be minimal)
    simple_prompts=$(echo "$prompts" | grep -c 'prompt="> "' || true)
    menu_prompts=$(echo "$prompts" | grep -c 'prompt="Choose an option' || true)
    
    # Should have both simple and menu prompts
    [ "$simple_prompts" -gt 0 ]
    [ "$menu_prompts" -gt 0 ]
}

@test "All headings use Title Case" {
    # Check various headings
    delete_heading=$(grep 'Delete a Radio Station' ../lib/delete_station.sh)
    lucky_heading=$(grep 'I Feel Lucky' ../lib/lucky.sh)
    
    # Verify they don't use all caps
    ! echo "$delete_heading" | grep -q "DELETE A RADIO STATION"
    ! echo "$lucky_heading" | grep -q "I FEEL LUCKY"
}

@test "No redundant text after search completes" {
    # Check that searching message is cleared
    result=$(grep -A 5 'greenprint "Searching ..."' ../lib/lib.sh)
    
    # Should have cleanup code after wget/curl (look for echo -ne)
    echo "$result" | grep -q 'Clear the'
}

@test "All clear commands come before headings" {
    # Check that pages clear screen before showing heading
    delete_clear=$(grep -B 1 'Delete a Radio Station' ../lib/delete_station.sh | grep 'clear')
    lucky_clear=$(grep -B 1 'I Feel Lucky' ../lib/lucky.sh | grep 'clear')
    
    [ -n "$delete_clear" ]
    [ -n "$lucky_clear" ]
}

@test "No double Main Menu entries in any menu" {
    # Check that menus don't have duplicate Main Menu entries
    list_menu=$(grep -A 10 'list_menu()' ../lib/list.sh | grep -c 'Main Menu')
    search_menu=$(grep -A 10 'search_menu()' ../lib/search.sh | grep -c 'Main Menu')
    
    # Each should have exactly 1 Main Menu entry in the options
    [ "$list_menu" -eq 1 ]
    [ "$search_menu" -eq 1 ]
}
