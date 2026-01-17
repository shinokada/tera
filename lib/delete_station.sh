#!/usr/bin/env bash

fn_delete() {
    TEMP_FILE="$TMP_PATH/radio_delete.json"
    _cleanup_tmp "$TEMP_FILE"
    clear
    cyanprint "$APP_NAME - Delete a Radio Station"
    echo
    touch "$TEMP_FILE"
    greenprint "Select a list to delete from."
    echo
    
    # Get all favorite lists
    LISTS=$(_fav_list)
    
    if [ -z "$LISTS" ]; then
        redprint "Lists are empty."
        cyanprint "Try $SCRIPT_NAME search"
        sleep 2
        menu
        return
    fi
    
    # Add Main Menu option and format for fzf
    MENU_OPTIONS="0) Main Menu"
    INDEX=1
    for list in $LISTS; do
        # Display "My Favorites" for myfavorites
        if [ "$list" = "My-favorites" ]; then
            DISPLAY_NAME="My Favorites"
        else
            DISPLAY_NAME="$list"
        fi
        MENU_OPTIONS="${MENU_OPTIONS}
${INDEX}) ${DISPLAY_NAME}"
        INDEX=$((INDEX + 1))
    done
    
    # Use fzf for list selection with arrow keys
    CHOICE=$(echo "$MENU_OPTIONS" | fzf --prompt="Choose a list (arrow keys to navigate): " --height=40% --reverse --no-info)
    
    # Check if user cancelled
    if [ -z "$CHOICE" ]; then
        menu
        return
    fi
    
    # Extract the number and list name
    LIST_NUM=$(echo "$CHOICE" | cut -d')' -f1)
    
    # Check if Main Menu was selected
    if [ "$LIST_NUM" = "0" ]; then
        menu
        return
    fi
    
    DISPLAY_NAME=$(echo "$CHOICE" | cut -d')' -f2- | sed 's/^ //')
    
    # Convert "My Favorites" display name back to "My-favorites" for file operations
    if [ "$DISPLAY_NAME" = "My Favorites" ]; then
        LIST="My-favorites"
    else
        LIST="$DISPLAY_NAME"
    fi
    
    clear
    cyanprint "$APP_NAME - Delete from $DISPLAY_NAME"
    echo
    
    # Get stations from the selected list
    STATIONS=$(_station_list "$LIST")
    
    if [ -z "$STATIONS" ]; then
        yellowprint "This list is empty."
        sleep 2
        menu
        return
    fi
    
    # Add Main Menu option
    STATIONS_WITH_MENU=$(printf "<< Main Menu >>\n%s" "$STATIONS")
    
    # Use fzf to select station with arrow keys
    SELECTION=$(echo "$STATIONS_WITH_MENU" | nl | fzf --prompt="Choose station to delete (arrow keys): " --height=40% --reverse --header="Delete from: $DISPLAY_NAME")
    
    # Check if user cancelled
    if [ -z "$SELECTION" ]; then
        menu
        return
    fi
    
    # Extract the selection text and number
    SELECTED_TEXT=$(echo "$SELECTION" | awk '{$1=""; print $0}' | sed 's/^ //')
    ANS=$(echo "$SELECTION" | awk '{print $1}')
    
    # Check if Main Menu was selected
    if [ "$SELECTED_TEXT" = "<< Main Menu >>" ]; then
        menu
        return
    fi
    
    # Adjust ANS to account for the Main Menu option (subtract 1)
    ANS=$((ANS - 1))
    
    # Since stations are sorted alphabetically, we need to find by name instead of index
    # Get the trimmed station name from the selection
    STATION_TO_DELETE=$(echo "$SELECTED_TEXT" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    
    # Show confirmation
    echo
    yellowprint "Are you sure you want to delete: $STATION_TO_DELETE"
    printf "Type 'yes' or 'y' to confirm, anything else to cancel: "
    read -r CONFIRM
    
    USER_CONFIRM=$(echo "$CONFIRM" | cut -c 1-1 | tr "[:lower:]" "[:upper:]")
    
    if [ "$USER_CONFIRM" = "Y" ]; then
        FAVLIST_PATH="${FAVORITE_PATH}/${LIST}.json"
        # Delete by matching the trimmed station name
        jq --arg name "$STATION_TO_DELETE" 'del(.[] | select(.name | gsub("^\\s+|\\s+$";"") == $name))' <"${FAVLIST_PATH}" >"$TEMP_FILE" && mv "$TEMP_FILE" "$FAVLIST_PATH"
        echo
        greenprint "Successfully deleted: $STATION_TO_DELETE"
        sleep 2
    else
        echo
        yellowprint "Deletion cancelled."
        sleep 1
    fi
    
    menu
}
