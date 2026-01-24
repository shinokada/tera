#!/usr/bin/env bats

# Integration tests for Gist Menu functionality

setup() {
    # Set up test environment
    export TEST_DIR="$BATS_TEST_DIRNAME/test_temp"
    export SCRIPT_DOT_DIR="$TEST_DIR/.config/tera"
    export FAVORITE_PATH="$SCRIPT_DOT_DIR/favorite"
    export GIST_METADATA_FILE="$SCRIPT_DOT_DIR/gist_metadata.json"
    export GIST_URL_FILE="$SCRIPT_DOT_DIR/gisturl"
    export APP_NAME="TERA"
    export GITHUB_TOKEN="fake_token_for_testing"
    
    # Create test directories
    mkdir -p "$SCRIPT_DOT_DIR"
    mkdir -p "$FAVORITE_PATH"
    
    # Create test favorite lists
    echo '[{"name":"Test Station 1","url":"http://example.com/1"}]' > "$FAVORITE_PATH/My-favorites.json"
    echo '[{"name":"Test Station 2","url":"http://example.com/2"}]' > "$FAVORITE_PATH/test-list.json"
    
    # Source required libraries
    source "$BATS_TEST_DIRNAME/../lib/lib.sh"
    source "$BATS_TEST_DIRNAME/../lib/gist_storage.sh"
}

teardown() {
    # Clean up test environment
    rm -rf "$TEST_DIR"
}

@test "gist metadata file is created on first use" {
    [ ! -f "$GIST_METADATA_FILE" ]
    
    init_gist_metadata
    
    [ -f "$GIST_METADATA_FILE" ]
}

@test "create_gist saves metadata after successful creation" {
    # Simulate a successful gist creation
    GIST_ID="test123abc"
    GIST_URL="https://gist.github.com/user/test123abc"
    
    save_gist_metadata "$GIST_ID" "$GIST_URL" "Terminal radio favorite lists"
    
    # Verify metadata was saved
    gist=$(get_gist_by_id "$GIST_ID")
    [ -n "$gist" ]
    
    saved_url=$(echo "$gist" | jq -r '.url')
    [ "$saved_url" = "$GIST_URL" ]
}

@test "list_my_gists shows correct count" {
    # Add test gists
    save_gist_metadata "gist1" "https://gist.github.com/user/gist1" "Gist 1"
    save_gist_metadata "gist2" "https://gist.github.com/user/gist2" "Gist 2"
    save_gist_metadata "gist3" "https://gist.github.com/user/gist3" "Gist 3"
    
    count=$(get_gist_count)
    [ "$count" -eq 3 ]
}

@test "recover_gist can select from saved gists" {
    # Add test gists
    save_gist_metadata "gist1" "https://gist.github.com/user/gist1" "Gist 1"
    save_gist_metadata "gist2" "https://gist.github.com/user/gist2" "Gist 2"
    
    # Verify we can retrieve them
    all_gists=$(get_all_gists)
    
    # Select first gist
    first_url=$(echo "$all_gists" | jq -r '.[0].url')
    [ "$first_url" = "https://gist.github.com/user/gist1" ]
    
    # Select second gist
    second_url=$(echo "$all_gists" | jq -r '.[1].url')
    [ "$second_url" = "https://gist.github.com/user/gist2" ]
}

@test "delete_gist removes from metadata" {
    # Add test gists
    save_gist_metadata "gist1" "https://gist.github.com/user/gist1" "Gist 1"
    save_gist_metadata "gist2" "https://gist.github.com/user/gist2" "Gist 2"
    
    initial_count=$(get_gist_count)
    [ "$initial_count" -eq 2 ]
    
    # Delete one gist
    delete_gist_metadata "gist1"
    
    final_count=$(get_gist_count)
    [ "$final_count" -eq 1 ]
    
    # Verify the correct gist was deleted
    remaining=$(get_gist_by_id "gist2")
    [ -n "$remaining" ]
    
    deleted=$(get_gist_by_id "gist1")
    [ -z "$deleted" ]
}

@test "gist menu shows gist count when gists exist" {
    save_gist_metadata "gist1" "https://gist.github.com/user/gist1" "Test Gist"
    
    count=$(get_gist_count)
    [ "$count" -gt 0 ]
}

@test "gist menu handles no gists gracefully" {
    count=$(get_gist_count)
    [ "$count" -eq 0 ]
}

@test "gist URLs are stored correctly" {
    GIST_URL="https://gist.github.com/user/abc123def456"
    save_gist_metadata "abc123def456" "$GIST_URL" "Test Gist"
    
    gist=$(get_gist_by_id "abc123def456")
    saved_url=$(echo "$gist" | jq -r '.url')
    
    [ "$saved_url" = "$GIST_URL" ]
}

@test "gist descriptions are stored correctly" {
    DESCRIPTION="My favorite radio stations - Jazz and Classical"
    save_gist_metadata "test123" "https://gist.github.com/user/test123" "$DESCRIPTION"
    
    gist=$(get_gist_by_id "test123")
    saved_desc=$(echo "$gist" | jq -r '.description')
    
    [ "$saved_desc" = "$DESCRIPTION" ]
}

@test "gist creation timestamp is recorded" {
    save_gist_metadata "test123" "https://gist.github.com/user/test123" "Test Gist"
    
    gist=$(get_gist_by_id "test123")
    created_at=$(echo "$gist" | jq -r '.created_at')
    
    [ -n "$created_at" ]
    [ "$created_at" != "null" ]
    
    # Verify timestamp format (ISO 8601)
    [[ "$created_at" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]
}

@test "multiple gists can coexist" {
    save_gist_metadata "gist1" "https://gist.github.com/user/gist1" "First Gist"
    save_gist_metadata "gist2" "https://gist.github.com/user/gist2" "Second Gist"
    save_gist_metadata "gist3" "https://gist.github.com/user/gist3" "Third Gist"
    
    count=$(get_gist_count)
    [ "$count" -eq 3 ]
    
    # Verify each gist exists independently
    gist1=$(get_gist_by_id "gist1")
    gist2=$(get_gist_by_id "gist2")
    gist3=$(get_gist_by_id "gist3")
    
    [ -n "$gist1" ]
    [ -n "$gist2" ]
    [ -n "$gist3" ]
}

@test "gist order is maintained" {
    save_gist_metadata "first" "https://gist.github.com/user/first" "First"
    save_gist_metadata "second" "https://gist.github.com/user/second" "Second"
    save_gist_metadata "third" "https://gist.github.com/user/third" "Third"
    
    all_gists=$(get_all_gists)
    
    first_id=$(echo "$all_gists" | jq -r '.[0].id')
    second_id=$(echo "$all_gists" | jq -r '.[1].id')
    third_id=$(echo "$all_gists" | jq -r '.[2].id')
    
    [ "$first_id" = "first" ]
    [ "$second_id" = "second" ]
    [ "$third_id" = "third" ]
}

@test "empty metadata file doesn't cause errors" {
    echo "[]" > "$GIST_METADATA_FILE"
    
    count=$(get_gist_count)
    [ "$count" -eq 0 ]
    
    all_gists=$(get_all_gists)
    [ "$all_gists" = "[]" ]
}

@test "corrupted metadata file is handled gracefully" {
    # Note: This test verifies behavior when file is malformed
    # In production, init_gist_metadata should be called first
    
    echo "invalid json" > "$GIST_METADATA_FILE"
    
    # This should fail gracefully when trying to read
    run get_gist_count
    [ "$status" -ne 0 ]
}

@test "gist IDs are unique identifiers" {
    save_gist_metadata "unique123" "https://gist.github.com/user/unique123" "Unique Gist"
    
    gist=$(get_gist_by_id "unique123")
    [ -n "$gist" ]
    
    # Try to get non-existent gist
    non_existent=$(get_gist_by_id "doesnotexist")
    [ -z "$non_existent" ]
}

@test "deleting all gists results in empty metadata" {
    save_gist_metadata "gist1" "https://gist.github.com/user/gist1" "Gist 1"
    save_gist_metadata "gist2" "https://gist.github.com/user/gist2" "Gist 2"
    
    delete_gist_metadata "gist1"
    delete_gist_metadata "gist2"
    
    count=$(get_gist_count)
    [ "$count" -eq 0 ]
    
    all_gists=$(get_all_gists)
    [ "$all_gists" = "[]" ]
}

@test "gist metadata persists across operations" {
    # Create
    save_gist_metadata "persist" "https://gist.github.com/user/persist" "Persistent Gist"
    
    # Read
    gist=$(get_gist_by_id "persist")
    [ -n "$gist" ]
    
    # Update
    update_gist_metadata "persist" "Updated Persistent Gist"
    
    # Verify update persisted
    updated_gist=$(get_gist_by_id "persist")
    desc=$(echo "$updated_gist" | jq -r '.description')
    [ "$desc" = "Updated Persistent Gist" ]
    
    # Verify we can still get all gists
    all_gists=$(get_all_gists)
    count=$(echo "$all_gists" | jq 'length')
    [ "$count" -eq 1 ]
}
