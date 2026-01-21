#!/usr/bin/env bats

# Tests for GitHub Token Management Functions

setup() {
    export TEST_DIR="$BATS_TEST_DIRNAME/test_temp"
    export SCRIPT_DOT_DIR="$TEST_DIR/.config/tera"
    export TOKENS_DIR="$SCRIPT_DOT_DIR/tokens"
    export GITHUB_TOKEN_FILE="$TOKENS_DIR/github_token"
    
    mkdir -p "$TOKENS_DIR"
    
    source "$BATS_TEST_DIRNAME/../lib/gist_storage.sh"
    source "$BATS_TEST_DIRNAME/../lib/lib.sh"
}

teardown() {
    rm -rf "$TEST_DIR"
}

@test "init_token_directory creates tokens directory" {
    rm -rf "$TOKENS_DIR"
    init_token_directory
    [ -d "$TOKENS_DIR" ]
}

@test "save_github_token saves token to file" {
    save_github_token "ghp_testtoken123456789"
    [ -f "$GITHUB_TOKEN_FILE" ]
}

@test "save_github_token overwrites existing token" {
    save_github_token "ghp_oldtoken123"
    save_github_token "ghp_newtoken456"
    
    content=$(cat "$GITHUB_TOKEN_FILE")
    [ "$content" = "ghp_newtoken456" ]
}

@test "save_github_token rejects empty token" {
    rm -f "$GITHUB_TOKEN_FILE"
    ! save_github_token ""
    [ ! -f "$GITHUB_TOKEN_FILE" ]
}

@test "load_github_token retrieves saved token" {
    echo "ghp_testtoken789" > "$GITHUB_TOKEN_FILE"
    
    result=$(load_github_token)
    [ "$result" = "ghp_testtoken789" ]
}

@test "load_github_token returns empty when file missing" {
    rm -f "$GITHUB_TOKEN_FILE"
    
    result=$(load_github_token)
    [ "$result" = "" ]
}

@test "has_github_token returns true when token exists" {
    echo "ghp_testtoken" > "$GITHUB_TOKEN_FILE"
    
    has_github_token
    [ $? -eq 0 ]
}

@test "delete_github_token removes token file" {
    echo "ghp_testtoken" > "$GITHUB_TOKEN_FILE"
    
    delete_github_token
    [ ! -f "$GITHUB_TOKEN_FILE" ]
}

@test "delete_github_token fails when file missing" {
    rm -f "$GITHUB_TOKEN_FILE"
    
    ! delete_github_token
}

@test "get_masked_token masks token correctly" {
    result=$(get_masked_token "ghp_abcdefghijklmnopqrstuvwxyz123456")
    
    [ -n "$result" ]
    case "$result" in
        ghp_*...*)
            true
            ;;
        *)
            false
            ;;
    esac
}

@test "get_masked_token shows last 4 characters" {
    result=$(get_masked_token "ghp_0123456789")
    
    case "$result" in
        *6789)
            true
            ;;
        *)
            false
            ;;
    esac
}

@test "get_masked_token handles short tokens" {
    result=$(get_masked_token "ghp_abc")
    [ -n "$result" ]
}

@test "save and load token preserves value" {
    TEST_TOKEN="ghp_workflowtest123456"
    
    save_github_token "$TEST_TOKEN"
    result=$(load_github_token)
    
    [ "$result" = "$TEST_TOKEN" ]
}

@test "has_github_token detects saved token" {
    save_github_token "ghp_workflow789"
    
    has_github_token
    [ $? -eq 0 ]
}

@test "token file contains no extra whitespace" {
    save_github_token "ghp_puretoken"
    
    file_size=$(wc -c < "$GITHUB_TOKEN_FILE")
    [ "$file_size" -le 60 ]
}

@test "token is stored on single line" {
    save_github_token "ghp_multiline_test"
    
    line_count=$(wc -l < "$GITHUB_TOKEN_FILE")
    [ "$line_count" -eq 1 ]
}

@test "init_token_directory is idempotent" {
    init_token_directory
    [ -d "$TOKENS_DIR" ]
    
    init_token_directory
    [ -d "$TOKENS_DIR" ]
}

@test "handle token with special characters" {
    SPECIAL_TOKEN="ghp_abc_123test"
    
    save_github_token "$SPECIAL_TOKEN"
    result=$(load_github_token)
    
    [ "$result" = "$SPECIAL_TOKEN" ]
}

@test "get_masked_token handles empty token gracefully" {
    result=$(get_masked_token "")
    
    [ -n "$result" ]
}

@test "multiple save operations only keep latest token" {
    save_github_token "ghp_first_token"
    save_github_token "ghp_second_token"
    save_github_token "ghp_third_token"
    
    result=$(load_github_token)
    [ "$result" = "ghp_third_token" ]
}

@test "token persistence across function calls" {
    save_github_token "ghp_persistent"
    
    has_github_token
    [ $? -eq 0 ]
    
    result=$(load_github_token)
    [ "$result" = "ghp_persistent" ]
    
    has_github_token
    [ $? -eq 0 ]
}

@test "delete followed by file check" {
    save_github_token "ghp_tobedeleted"
    delete_github_token
    
    [ ! -f "$GITHUB_TOKEN_FILE" ]
}
