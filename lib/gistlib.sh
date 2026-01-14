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
        yellowprint "Or run: cp ${script_dir}/.env.example ${script_dir}/.env"
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
        filename=$(basename "$file")
        # Read file content and escape for JSON
        content=$(jq -Rs . < "$file")
        
        if [ "$FIRST_FILE" = true ]; then
            FIRST_FILE=false
        else
            JSON_PAYLOAD="${JSON_PAYLOAD},"
        fi
        
        JSON_PAYLOAD="${JSON_PAYLOAD}\"${filename}\":{\"content\":${content}}"
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
    ERROR_MSG=$(echo "$RESPONSE" | jq -r '.message // empty')
    
    if [ -n "$GIST_URL" ] && [ "$GIST_URL" != "null" ]; then
        greenprint "✓ Successfully created a secret Gist!"
        echo
        cyanprint "Gist URL: $GIST_URL"
        echo "$GIST_URL" > "$GIST_URL_FILE"
        echo
        greenprint "Opening in browser..."
        python3 -m webbrowser "$GIST_URL" 2>/dev/null || true
        echo
        read -p "Press Enter to return to menu..."
        gist_menu
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

recover_gist() {
    clear
    cyanprint "$APP_NAME - Recover Favorites from a Gist"
    echo
    yellowprint "Type '0' to go back to Gist Menu"
    printf "What is your Gist URL? "
    read -r gist_url
    
    # Check for navigation
    if [ "$gist_url" = "0" ] || [ -z "$gist_url" ]; then
        gist_menu
        return
    fi
    
    echo
    greenprint "Cloning gist..."
    cd "$FAVORITE_PATH" || {
        redprint "Error: Could not access favorites directory."
        sleep 2
        gist_menu
        return
    }
    
    if git clone "$gist_url" 2>/dev/null; then
        # find the last from the path
        gist_dir=${gist_url##*/}
        
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
    
    echo
    read -p "Press Enter to return to menu..."
    gist_menu
}

gist_menu() {
    clear
    cyanprint "$APP_NAME GIST MENU"
    echo
    
    MENU_OPTIONS="0) Main Menu
1) Create a gist
2) Recover favorites from a gist
3) Exit"
    
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
        recover_gist
        ;;
    3)
        yellowprint "Bye-bye."
        exit 0
        ;;
    *)
        redprint "Wrong option."
        gist_menu
        ;;
    esac
}
