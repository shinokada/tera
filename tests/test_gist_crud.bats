#!/usr/bin/env bats

# Tests for Gist CRUD operations

setup() {
    # Set up test environment
    export TEST_DIR="$BATS_TEST_DIRNAME/test_temp"
    export SCRIPT_DOT_DIR="$TEST_DIR/.config/tera"
    export FAVORITE_PATH="$SCRIPT_DOT_DIR/favorite"
    export GIST_METADATA_FILE="$SCRIPT_DOT_DIR/gist_metadata.json"
    export APP_NAME="TERA"
    
    # Create test directories
    mkdir -p "$SCRIPT_DOT_DIR"
    mkdir -p "$FAVORITE_PATH"
    
    # Source the gist storage library
    source "$BATS_TEST_DIRNAME/../lib/gist_storage.sh"
    
    # Source color functions from lib.sh
    source "$BATS_TEST_DIRNAME/../lib/lib.sh"
}

teardown() {
    # Clean up test environment
    rm -rf "$TEST_DIR"
}

@test "init_gist_metadata creates empty JSON array" {
    init_gist_metadata
    
    [ -f "$GIST_METADATA_FILE" ]
    
    result=$(cat "$GIST_METADATA_FILE")
    [ "$result" = "[]" ]
}

@test "init_gist_metadata doesn't overwrite existing file" {
    # Create a file with content
    echo '[{"id":"test"}]' > "$GIST_METADATA_FILE"
    
    init_gist_metadata
    
    result=$(cat "$GIST_METADATA_FILE")
    [ "$result" = '[{"id":"test"}]' ]
}

@test "save_gist_metadata adds new gist" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "Test gist"
    
    [ -f "$GIST_METADATA_FILE" ]
    
    # Check if gist was added
    gist_count=$(jq 'length' "$GIST_METADATA_FILE")
    [ "$gist_count" -eq 1 ]
    
    # Check gist properties
    gist_id=$(jq -r '.[0].id' "$GIST_METADATA_FILE")
    [ "$gist_id" = "abc123" ]
    
    gist_url=$(jq -r '.[0].url' "$GIST_METADATA_FILE")
    [ "$gist_url" = "https://gist.github.com/user/abc123" ]
    
    gist_desc=$(jq -r '.[0].description' "$GIST_METADATA_FILE")
    [ "$gist_desc" = "Test gist" ]
}

@test "save_gist_metadata adds multiple gists" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "First gist"
    save_gist_metadata "def456" "https://gist.github.com/user/def456" "Second gist"
    save_gist_metadata "ghi789" "https://gist.github.com/user/ghi789" "Third gist"
    
    gist_count=$(jq 'length' "$GIST_METADATA_FILE")
    [ "$gist_count" -eq 3 ]
}

@test "get_all_gists returns all gists" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "First gist"
    save_gist_metadata "def456" "https://gist.github.com/user/def456" "Second gist"
    
    result=$(get_all_gists)
    
    # Check if result is valid JSON array
    gist_count=$(echo "$result" | jq 'length')
    [ "$gist_count" -eq 2 ]
}

@test "get_all_gists returns empty array when no gists" {
    result=$(get_all_gists)
    
    [ "$result" = "[]" ]
}

@test "get_gist_by_id finds correct gist" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "First gist"
    save_gist_metadata "def456" "https://gist.github.com/user/def456" "Second gist"
    
    result=$(get_gist_by_id "def456")
    
    gist_id=$(echo "$result" | jq -r '.id')
    [ "$gist_id" = "def456" ]
    
    gist_desc=$(echo "$result" | jq -r '.description')
    [ "$gist_desc" = "Second gist" ]
}

@test "get_gist_by_id returns empty when gist doesn't exist" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "First gist"
    
    result=$(get_gist_by_id "nonexistent")
    
    [ -z "$result" ]
}

@test "update_gist_metadata updates description" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "Original description"
    
    update_gist_metadata "abc123" "Updated description"
    
    gist_desc=$(jq -r '.[0].description' "$GIST_METADATA_FILE")
    [ "$gist_desc" = "Updated description" ]
}

@test "update_gist_metadata updates updated_at timestamp" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "Test gist"
    
    # Get original timestamps
    original_created=$(jq -r '.[0].created_at' "$GIST_METADATA_FILE")
    original_updated=$(jq -r '.[0].updated_at' "$GIST_METADATA_FILE")
    
    # Wait a moment to ensure timestamp difference
    sleep 1
    
    update_gist_metadata "abc123" "Updated description"
    
    # Check that created_at didn't change
    new_created=$(jq -r '.[0].created_at' "$GIST_METADATA_FILE")
    [ "$new_created" = "$original_created" ]
    
    # Check that updated_at changed
    new_updated=$(jq -r '.[0].updated_at' "$GIST_METADATA_FILE")
    [ "$new_updated" != "$original_updated" ]
}

@test "update_gist_metadata doesn't affect other gists" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "First gist"
    save_gist_metadata "def456" "https://gist.github.com/user/def456" "Second gist"
    
    update_gist_metadata "abc123" "Updated first gist"
    
    # Check first gist was updated
    first_desc=$(jq -r '.[] | select(.id == "abc123") | .description' "$GIST_METADATA_FILE")
    [ "$first_desc" = "Updated first gist" ]
    
    # Check second gist wasn't changed
    second_desc=$(jq -r '.[] | select(.id == "def456") | .description' "$GIST_METADATA_FILE")
    [ "$second_desc" = "Second gist" ]
}

@test "delete_gist_metadata removes correct gist" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "First gist"
    save_gist_metadata "def456" "https://gist.github.com/user/def456" "Second gist"
    save_gist_metadata "ghi789" "https://gist.github.com/user/ghi789" "Third gist"
    
    delete_gist_metadata "def456"
    
    # Check count
    gist_count=$(jq 'length' "$GIST_METADATA_FILE")
    [ "$gist_count" -eq 2 ]
    
    # Check the deleted gist is gone
    result=$(get_gist_by_id "def456")
    [ -z "$result" ]
    
    # Check other gists still exist
    first=$(get_gist_by_id "abc123")
    [ -n "$first" ]
    
    third=$(get_gist_by_id "ghi789")
    [ -n "$third" ]
}

@test "delete_gist_metadata handles nonexistent gist gracefully" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "First gist"
    
    delete_gist_metadata "nonexistent"
    
    # Check original gist still exists
    gist_count=$(jq 'length' "$GIST_METADATA_FILE")
    [ "$gist_count" -eq 1 ]
}

@test "get_gist_count returns correct count" {
    result=$(get_gist_count)
    [ "$result" -eq 0 ]
    
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "First gist"
    result=$(get_gist_count)
    [ "$result" -eq 1 ]
    
    save_gist_metadata "def456" "https://gist.github.com/user/def456" "Second gist"
    result=$(get_gist_count)
    [ "$result" -eq 2 ]
    
    delete_gist_metadata "abc123"
    result=$(get_gist_count)
    [ "$result" -eq 1 ]
}

@test "gist metadata includes all required fields" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "Test gist"
    
    gist=$(get_gist_by_id "abc123")
    
    # Check id field
    id=$(echo "$gist" | jq -r '.id')
    [ -n "$id" ]
    [ "$id" != "null" ]
    
    # Check url field
    url=$(echo "$gist" | jq -r '.url')
    [ -n "$url" ]
    [ "$url" != "null" ]
    
    # Check description field
    desc=$(echo "$gist" | jq -r '.description')
    [ -n "$desc" ]
    [ "$desc" != "null" ]
    
    # Check created_at field
    created=$(echo "$gist" | jq -r '.created_at')
    [ -n "$created" ]
    [ "$created" != "null" ]
    
    # Check updated_at field
    updated=$(echo "$gist" | jq -r '.updated_at')
    [ -n "$updated" ]
    [ "$updated" != "null" ]
}

@test "timestamps are in ISO 8601 format" {
    save_gist_metadata "abc123" "https://gist.github.com/user/abc123" "Test gist"
    
    created=$(jq -r '.[0].created_at' "$GIST_METADATA_FILE")
    
    # Check format: YYYY-MM-DDTHH:MM:SSZ
    [[ "$created" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]
}

@test "multiple operations maintain data integrity" {
    # Add gists
    save_gist_metadata "gist1" "https://gist.github.com/user/gist1" "Gist 1"
    save_gist_metadata "gist2" "https://gist.github.com/user/gist2" "Gist 2"
    save_gist_metadata "gist3" "https://gist.github.com/user/gist3" "Gist 3"
    
    # Update one
    update_gist_metadata "gist2" "Updated Gist 2"
    
    # Delete one
    delete_gist_metadata "gist1"
    
    # Add another
    save_gist_metadata "gist4" "https://gist.github.com/user/gist4" "Gist 4"
    
    # Verify final state
    count=$(get_gist_count)
    [ "$count" -eq 3 ]
    
    # Verify correct gists exist
    gist2=$(get_gist_by_id "gist2")
    desc2=$(echo "$gist2" | jq -r '.description')
    [ "$desc2" = "Updated Gist 2" ]
    
    gist3=$(get_gist_by_id "gist3")
    [ -n "$gist3" ]
    
    gist4=$(get_gist_by_id "gist4")
    [ -n "$gist4" ]
    
    # Verify deleted gist is gone
    gist1=$(get_gist_by_id "gist1")
    [ -z "$gist1" ]
}
