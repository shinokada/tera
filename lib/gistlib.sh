#!/usr/bin/env bash

GIST_URL_FILE="$SCRIPT_DOT_DIR/gisturl"

create_gist() {
    clear
    cyanprint "$APP_NAME - Create a Gist"
    echo
    
    # Check if GitHub token is available
    if [ -z "$GITHUB_TOKEN" ]; then
        redprint "Error: GitHub token not found!"
        echo
        yellowprint "To create gists, you need to set up a GitHub token:"
        echo "1. Copy .env.example to .env in the tera directory"
        echo "2. Get a token from https://github.com/settings/tokens"
        echo "3. Make sure to select the 'gist' scope"
        echo "4. Paste your token in the .env file"
        echo
        yellowprint "Or run: cp .env.example .env"
        yellowprint "Then edit the .env file with your token"
        echo
        read -p "Press Enter to return to menu..."
        gist_menu
        return
    fi
    
    # Get favorite lists
    FAV=$(_list_intro 2>/dev/null) || true
    
    # Check if there are any lists
    if [ -z "$FAV" ]; then
        redprint "No favorite lists found!"
        echo
        yellowprint "Create some lists and add stations first."
        echo "Try: Search → Save stations to lists"
        echo
        read -p "Press Enter to return to menu..."
        gist_menu
        return
    fi
    
    ARR=()
    # add fav list to ARR array
    for file in $FAV; do
        if [ -f "${FAVORITE_PATH}/${file}.json" ]; then
            ARR+=("${FAVORITE_PATH}/${file}.json")
        fi
    done
    
    # Check if we have any files to upload
    if [ ${#ARR[@]} -eq 0 ]; then
        redprint "No valid list files found!"
        echo
        read -p "Press Enter to return to menu..."
        gist_menu
        return
    fi
    
    greenprint "Creating gist using GitHub API..."
    greenprint "Uploading ${#ARR[@]} list file(s)..."
    echo
    
    # Build JSON payload for GitHub API
    # Start with basic gist structure
    JSON_PAYLOAD='{"description":"Terminal radio favorite lists","public":false,"files":{'
    
    FIRST_FILE=true
    for file in "${ARR[@]}"; do
        # Escape filename properly for JSON
        filename=$(basename "$file" | jq -Rs '.[:-1]')  # Remove trailing newline from jq output
        # Read file content and escape for JSON
        content=$(jq -Rs . < "$file")
        
        if [ "$FIRST_FILE" = true ]; then
            FIRST_FILE=false
        else
            JSON_PAYLOAD="${JSON_PAYLOAD},"
        fi
        
        JSON_PAYLOAD="${JSON_PAYLOAD}${filename}:{\"content\":${content}}"
    done
    
    JSON_PAYLOAD="${JSON_PAYLOAD}}}"
    
    # Create gist using curl and GitHub API
    RESPONSE=$(curl -s -X POST \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer ${GITHUB_TOKEN}" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        https://api.github.com/gists \
        -d "$JSON_PAYLOAD")
    
    # Check if gist was created successfully
    GIST_URL=$(echo "$RESPONSE" | jq -r '.html_url // empty')
    GIST_ID=$(echo "$RESPONSE" | jq -r '.id // empty')
    ERROR_MSG=$(echo "$RESPONSE" | jq -r '.message // empty')
    
    if [ -n "$GIST_URL" ] && [ "$GIST_URL" != "null" ]; then
        greenprint "✓ Successfully created a secret Gist!"
        echo
        cyanprint "Gist URL: $GIST_URL"
        echo "$GIST_URL" > "$GIST_URL_FILE"
        
        # Save gist metadata
        save_gist_metadata "$GIST_ID" "$GIST_URL" "Terminal radio favorite lists"
        
        echo
        greenprint "Opening in browser..."
        python3 -m webbrowser "$GIST_URL" 2>/dev/null || true
        echo
        read -p "Press Enter to return to menu..."
        gist_menu
        return
    else
        redprint "✗ Failed to create gist!"
        echo
        if [ -n "$ERROR_MSG" ]; then
            redprint "Error: $ERROR_MSG"
        else
            redprint "Unknown error occurred"
            yellowprint "Response: $RESPONSE"
        fi
        echo
        yellowprint "Common issues:"
        echo "1. Check if your GITHUB_TOKEN in .env is valid"
        echo "2. Verify the token has 'gist' scope"
        echo "3. Make sure you have internet connection"
        echo "4. Token might be expired - generate a new one"
        echo
        read -p "Press Enter to return to menu..."
        gist_menu
        return
    fi
}

# List all saved gists
list_my_gists() {
    clear
    cyanprint "$APP_NAME - My Gists"
    echo
    
    local gist_count=$(get_gist_count)
    
    if [ "$gist_count" -eq 0 ]; then
        yellowprint "You haven't created any gists yet."
        echo
        yellowprint "Create a gist from the Gist Menu to save your favorite lists."
        echo
        read -p "Press Enter to return to menu..."
        gist_menu
        return
    fi
    
    greenprint "You have $gist_count gist(s):"
    echo
    cyanprint "$(printf "%-50s | %s" "Description" "Created")"
    echo "$(printf '%.0s-' {1..80})"
    
    # Get all gists and display them
    local gists=$(get_all_gists)
    local index=0
    
    # Store gist info for selection
    declare -a GIST_IDS
    declare -a GIST_URLS
    declare -a GIST_DESCS
    
    echo "$gists" | jq -c '.[]' | while IFS= read -r gist; do
        index=$((index + 1))
        local id=$(echo "$gist" | jq -r '.id')
        local url=$(echo "$gist" | jq -r '.url')
        local desc=$(echo "$gist" | jq -r '.description')
        local created=$(echo "$gist" | jq -r '.created_at')
        
        # Format date
        local created_date=$(date -d "$created" "+%Y-%m-%d %H:%M" 2>/dev/null || echo "$created")
        
        printf "%2d) %-47s | %s\n" "$index" "$desc" "$created_date"
        
        GIST_IDS+=("$id")
        GIST_URLS+=("$url")
        GIST_DESCS+=("$desc")
    done
    
    echo
    echo "$(printf '%.0s-' {1..80})"
    echo
    yellowprint "Type '0' to go back to Gist Menu"
    printf "Enter gist number to open in browser (or press Enter to go back): "
    read -r choice
    
    # Navigation
    if [ "$choice" = "0" ] || [ -z "$choice" ]; then
        gist_menu
        return
    fi
    
    # Validate choice
    if ! [[ "$choice" =~ ^[0-9]+$ ]] || [ "$choice" -lt 1 ] || [ "$choice" -gt "$gist_count" ]; then
        redprint "Invalid choice."
        sleep 1
        list_my_gists
        return
    fi
    
    # Get the selected gist URL
    local selected_index=$((choice - 1))
    local selected_gist=$(echo "$gists" | jq -r ".[$selected_index]")
    local selected_url=$(echo "$selected_gist" | jq -r '.url')
    
    if [ -n "$selected_url" ] && [ "$selected_url" != "null" ]; then
        greenprint "Opening gist in browser..."
        python3 -m webbrowser "$selected_url" 2>/dev/null || true
    fi
    
    echo
    read -p "Press Enter to return to list..."
    list_my_gists
}

# Delete a gist (both from GitHub and local metadata)
delete_gist() {
    clear
    cyanprint "$APP_NAME - Delete Gist"
    echo
    
    # Check if GitHub token is available
    if [ -z "$GITHUB_TOKEN" ]; then
        redprint "Error: GitHub token not found!"
        echo
        yellowprint "To delete gists, you need a GitHub token configured."
        echo
        read -p "Press Enter to return to menu..."
        gist_menu
        return
    fi
    
    local gist_count=$(get_gist_count)
    
    if [ "$gist_count" -eq 0 ]; then
        yellowprint "You don't have any gists to delete."
        echo
        read -p "Press Enter to return to menu..."
        gist_menu
        return
    fi
    
    greenprint "Your gists:"
    echo
    cyanprint "$(printf "%-50s | %s" "Description" "Created")"
    echo "$(printf '%.0s-' {1..80})"
    
    # Get all gists and display them
    local gists=$(get_all_gists)
    local index=0
    
    echo "$gists" | jq -c '.[]' | while IFS= read -r gist; do
        index=$((index + 1))
        local desc=$(echo "$gist" | jq -r '.description')
        local created=$(echo "$gist" | jq -r '.created_at')
        local created_date=$(date -d "$created" "+%Y-%m-%d %H:%M" 2>/dev/null || echo "$created")
        
        printf "%2d) %-47s | %s\n" "$index" "$desc" "$created_date"
    done
    
    echo
    echo "$(printf '%.0s-' {1..80})"
    echo
    yellowprint "Type '0' to go back to Gist Menu"
    printf "Enter gist number to delete (or press Enter to cancel): "
    read -r choice
    
    # Navigation
    if [ "$choice" = "0" ] || [ -z "$choice" ]; then
        gist_menu
        return
    fi
    
    # Validate choice
    if ! [[ "$choice" =~ ^[0-9]+$ ]] || [ "$choice" -lt 1 ] || [ "$choice" -gt "$gist_count" ]; then
        redprint "Invalid choice."
        sleep 1
        delete_gist
        return
    fi
    
    # Get the selected gist
    local selected_index=$((choice - 1))
    local selected_gist=$(echo "$gists" | jq -r ".[$selected_index]")
    local gist_id=$(echo "$selected_gist" | jq -r '.id')
    local gist_desc=$(echo "$selected_gist" | jq -r '.description')
    
    # Confirm deletion
    echo
    yellowprint "Are you sure you want to delete this gist?"
    cyanprint "Description: $gist_desc"
    echo
    printf "Type 'yes' to confirm deletion (or anything else to cancel): "
    read -r confirm
    
    if [ "$confirm" != "yes" ]; then
        yellowprint "Deletion cancelled."
        sleep 1
        delete_gist
        return
    fi
    
    # Delete from GitHub
    greenprint "Deleting gist from GitHub..."
    RESPONSE=$(curl -s -X DELETE \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer ${GITHUB_TOKEN}" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "https://api.github.com/gists/$gist_id")
    
    # Check response (DELETE returns 204 No Content on success)
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
        -X DELETE \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer ${GITHUB_TOKEN}" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "https://api.github.com/gists/$gist_id")
    
    if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "404" ]; then
        # Delete from local metadata
        delete_gist_metadata "$gist_id"
        greenprint "✓ Gist deleted successfully!"
    else
        redprint "✗ Failed to delete gist from GitHub (HTTP $HTTP_CODE)"
        echo
        yellowprint "The gist will be removed from your local list anyway."
        delete_gist_metadata "$gist_id"
    fi
    
    echo
    read -p "Press Enter to continue..."
    delete_gist
}

recover_gist() {
    clear
    cyanprint "$APP_NAME - Recover Favorites from a Gist"
    echo
    
    # Show saved gists if any exist
    local gist_count=$(get_gist_count)
    if [ "$gist_count" -gt 0 ]; then
        greenprint "Your saved gists:"
        echo
        
        local gists=$(get_all_gists)
        local index=0
        
        echo "$gists" | jq -c '.[]' | while IFS= read -r gist; do
            index=$((index + 1))
            local url=$(echo "$gist" | jq -r '.url')
            local desc=$(echo "$gist" | jq -r '.description')
            local created=$(echo "$gist" | jq -r '.created_at')
            local created_date=$(date -d "$created" "+%Y-%m-%d %H:%M" 2>/dev/null || echo "$created")
            
            printf "%2d) %s (%s)\n" "$index" "$desc" "$created_date"
        done
        
        echo
        echo "$(printf '%.0s-' {1..60})"
        echo
        yellowprint "You can select a gist number above, or enter a gist URL manually."
        echo
    fi
    
    yellowprint "Type '0' to go back to Gist Menu"
    printf "Enter gist number or URL: "
    read -r gist_input
    
    # Check for navigation
    if [ "$gist_input" = "0" ] || [ -z "$gist_input" ]; then
        gist_menu
        return
    fi
    
    local gist_url=""
    
    # Check if input is a number (selecting from list)
    if [[ "$gist_input" =~ ^[0-9]+$ ]] && [ "$gist_count" -gt 0 ]; then
        if [ "$gist_input" -ge 1 ] && [ "$gist_input" -le "$gist_count" ]; then
            local selected_index=$((gist_input - 1))
            local gists=$(get_all_gists)
            gist_url=$(echo "$gists" | jq -r ".[$selected_index].url")
        else
            redprint "Invalid gist number."
            sleep 1
            recover_gist
            return
        fi
    else
        # Treat as URL
        gist_url="$gist_input"
    fi
    
    echo
    greenprint "Cloning gist..."
    pushd "$FAVORITE_PATH" > /dev/null || {
        redprint "Error: Could not access favorites directory."
        sleep 2
        gist_menu
        return
    }
    
    if git clone "$gist_url" 2>/dev/null; then
        # find the last from the path
        gist_dir=${gist_url%/}      # Remove trailing slash if present
        gist_dir=${gist_dir##*/}    # Extract last path segment
        
        if [ -z "$gist_dir" ] || [ ! -d "$FAVORITE_PATH/$gist_dir" ]; then
            popd > /dev/null
            redprint "Error: Could not determine gist directory."
            read -p "Press Enter to return to menu..."
            gist_menu
            return
        fi
        
        # Count how many files we're moving
        file_count=$(find "$FAVORITE_PATH/$gist_dir" -name "*.json" -type f 2>/dev/null | wc -l)
        
        if [ "$file_count" -gt 0 ]; then
            mv "$FAVORITE_PATH"/"$gist_dir"/*.json "$FAVORITE_PATH" 2>/dev/null || true
            rm -rf "./$gist_dir"
            greenprint "✓ Successfully downloaded $file_count list(s)!"
        else
            rm -rf "./$gist_dir"
            yellowprint "No JSON files found in the gist."
        fi
    else
        redprint "✗ Failed to clone gist."
        echo
        yellowprint "Make sure:"
        echo "1. The URL is correct"
        echo "2. The gist exists and is accessible"
        echo "3. You have internet connection"
    fi
    
    popd > /dev/null
    echo
    read -p "Press Enter to return to menu..."
    gist_menu
}

gist_menu() {
    clear
    cyanprint "$APP_NAME GIST MENU"
    echo
    
    # Show gist count if any exist
    local gist_count=$(get_gist_count)
    if [ "$gist_count" -gt 0 ]; then
        greenprint "You have $gist_count saved gist(s)"
        echo
    fi
    
    MENU_OPTIONS="0) Main Menu
1) Create a gist
2) My Gists
3) Recover favorites from a gist
4) Delete a gist
5) Exit"
    
    CHOICE=$(echo "$MENU_OPTIONS" | fzf --prompt="Choose an option (arrow keys to navigate): " --height=40% --reverse --no-info)
    
    if [ -z "$CHOICE" ]; then
        menu
        return
    fi
    
    ans=$(echo "$CHOICE" | cut -d')' -f1)
    
    case $ans in
    0)
        menu
        ;;
    1)
        create_gist
        ;;
    2)
        list_my_gists
        ;;
    3)
        recover_gist
        ;;
    4)
        delete_gist
        ;;
    5)
        yellowprint "Bye-bye."
        exit 0
        ;;
    *)
        redprint "Wrong option."
        gist_menu
        ;;
    esac
}
