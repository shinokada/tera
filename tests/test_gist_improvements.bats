#!/usr/bin/env bats

# Test gist function improvements
# Verifies proper return statements and directory handling

@test "create_gist has return after successful gist creation" {
    # Check that the success path has a return statement
    # Look for the success message and verify return comes after gist_menu
    awk '/Successfully created a secret Gist/,/^[[:space:]]*else/' ../lib/gistlib.sh | grep -A1 'gist_menu' | grep -q 'return'
}

@test "create_gist has return after gist creation failure" {
    # Check that the failure path has a return statement
    # The failure message is "Failed to create gist", check there's gist_menu + return nearby
    grep -A20 'Failed to create gist' ../lib/gistlib.sh | grep -A1 'gist_menu' | tail -2 | grep -q 'return'
}

@test "recover_gist uses pushd instead of cd" {
    # Check that pushd is used to change directory
    grep -q 'pushd "$FAVORITE_PATH"' ../lib/gistlib.sh
}

@test "recover_gist restores directory with popd" {
    # Check that popd is called to restore directory
    grep -q 'popd > /dev/null' ../lib/gistlib.sh
}

@test "recover_gist popd comes before final gist_menu call" {
    # Verify that popd happens before returning to menu
    popd_line=$(grep -n 'popd > /dev/null' ../lib/gistlib.sh | tail -n1 | cut -d: -f1)
    final_menu_line=$(grep -n 'read -p.*Press Enter to return to menu' ../lib/gistlib.sh | tail -n1 | cut -d: -f1)
    
    # popd should come before the final menu prompt
    [ "$popd_line" -lt "$final_menu_line" ]
}

@test "recover_gist popd on error path when gist_dir determination fails" {
    # Check that popd is called even on error path
    # Look for the error about determining gist directory and check popd comes before it
    awk '/if.*gist_dir.*FAVORITE_PATH/,/fi/' ../lib/gistlib.sh | grep -B1 'Could not determine gist directory' | grep -q 'popd'
}

@test "all gist_menu calls followed by return in create_gist" {
    # Count gist_menu calls in create_gist function
    menu_calls=$(awk '/^create_gist\(\)/,/^}/' ../lib/gistlib.sh | grep -c 'gist_menu')
    
    # Count return statements after gist_menu in create_gist
    returns=$(awk '/^create_gist\(\)/,/^}/' ../lib/gistlib.sh | grep -A1 'gist_menu' | grep -c 'return')
    
    # Every gist_menu call should have a return
    [ "$menu_calls" -eq "$returns" ]
}
