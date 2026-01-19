#!/usr/bin/env bash

# Gist metadata storage system
# Stores gist metadata in ~/.config/tera/gist_metadata.json

GIST_METADATA_FILE="$SCRIPT_DOT_DIR/gist_metadata.json"

# Initialize gist metadata file if it doesn't exist
init_gist_metadata() {
    if [ ! -f "$GIST_METADATA_FILE" ]; then
        mkdir -p "$(dirname "$GIST_METADATA_FILE")"
        echo "[]" > "$GIST_METADATA_FILE"
    fi
}

# Save gist metadata
# Usage: save_gist_metadata <gist_id> <gist_url> <description>
save_gist_metadata() {
    local gist_id="$1"
    local gist_url="$2"
    local description="$3"
    local created_at=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    init_gist_metadata
    
    # Check if gist already exists
    if [ -n "$(get_gist_by_id "$gist_id")" ]; then
        update_gist_metadata "$gist_id" "$description"
        return
    fi
    
    # Create new gist entry
    local new_entry=$(jq -n \
        --arg id "$gist_id" \
        --arg url "$gist_url" \
        --arg desc "$description" \
        --arg created "$created_at" \
        '{
            id: $id,
            url: $url,
            description: $desc,
            created_at: $created,
            updated_at: $created
        }')
    
    # Add to metadata file
    jq --argjson entry "$new_entry" '. += [$entry]' "$GIST_METADATA_FILE" > "${GIST_METADATA_FILE}.tmp"
    mv "${GIST_METADATA_FILE}.tmp" "$GIST_METADATA_FILE"
}

# Get all gists
# Returns JSON array of gist metadata
get_all_gists() {
    init_gist_metadata
    cat "$GIST_METADATA_FILE"
}

# Get gist by ID
# Usage: get_gist_by_id <gist_id>
get_gist_by_id() {
    local gist_id="$1"
    init_gist_metadata
    jq --arg id "$gist_id" '.[] | select(.id == $id)' "$GIST_METADATA_FILE"
}

# Update gist metadata
# Usage: update_gist_metadata <gist_id> <description>
update_gist_metadata() {
    local gist_id="$1"
    local description="$2"
    local updated_at=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    init_gist_metadata
    
    # Update the gist entry
    jq --arg id "$gist_id" \
       --arg desc "$description" \
       --arg updated "$updated_at" \
       'map(if .id == $id then .description = $desc | .updated_at = $updated else . end)' \
       "$GIST_METADATA_FILE" > "${GIST_METADATA_FILE}.tmp"
    mv "${GIST_METADATA_FILE}.tmp" "$GIST_METADATA_FILE"
}

# Delete gist metadata
# Usage: delete_gist_metadata <gist_id>
delete_gist_metadata() {
    local gist_id="$1"
    init_gist_metadata
    
    jq --arg id "$gist_id" 'map(select(.id != $id))' "$GIST_METADATA_FILE" > "${GIST_METADATA_FILE}.tmp"
    mv "${GIST_METADATA_FILE}.tmp" "$GIST_METADATA_FILE"
}

# Get gist count
get_gist_count() {
    init_gist_metadata
    jq 'length' "$GIST_METADATA_FILE"
}

# Format gist for display
# Usage: format_gist_display <gist_json>
format_gist_display() {
    local gist_json="$1"
    local description=$(echo "$gist_json" | jq -r '.description')
    local created=$(echo "$gist_json" | jq -r '.created_at')
    
    # Format created date to be more readable (cross-platform compatible)
    local created_date
    if date -d "$created" "+%Y-%m-%d %H:%M" >/dev/null 2>&1; then
        # GNU date (Linux)
        created_date=$(date -d "$created" "+%Y-%m-%d %H:%M")
    elif date -j -f "%Y-%m-%dT%H:%M:%SZ" "$created" "+%Y-%m-%d %H:%M" >/dev/null 2>&1; then
        # BSD date (macOS)
        created_date=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$created" "+%Y-%m-%d %H:%M")
    else
        created_date="$created"
    fi
    
    printf "%-50s | %s\n" "$description" "$created_date"
}
