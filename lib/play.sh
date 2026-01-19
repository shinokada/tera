#!/usr/bin/env bash

# # opens a favorite list and select number to play a station
# tera play
fn_play() {
    # check if a list is empty
    lists=$(_fav_list)
    # lists=""
    # echo "$lists"
    clear
    cyanprint "$APP_NAME - Play from My List"
    if [ -z "$lists" ]; then
        redprint "Lists are empty."
        cyanprint "Try $SCRIPT_NAME search"
    fi
    echo
    
    # Add Main Menu option
    lists_with_menu=$(echo "$lists" | tr ' ' '\n' | sed '1i<< Main Menu >>')
    
    LIST=$(echo "$lists_with_menu" | fzf --prompt="> " --header="$APP_NAME - Play from My List" --header-first --height=40% --reverse)
    
    # Check if user cancelled or Main Menu selected
    if [ -z "$LIST" ] || [ "$LIST" = "<< Main Menu >>" ]; then
        menu
        return
    fi

    # read -rp "Type a list number.   " LIST
    if [ -n "$LIST" ]; then
        echo
        STATIONS=$(_station_list "$LIST")
        
        # Add Main Menu option
        STATIONS_WITH_MENU=$(printf "<< Main Menu >>\n%s" "$STATIONS")
        
        # Use fzf to select station
        SELECTION=$(echo "$STATIONS_WITH_MENU" | nl | fzf --prompt="> " --header="$APP_NAME - Play from $LIST" --header-first --height=40% --reverse)
        
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
        
        LIST_PATH="$FAVORITE_PATH/$LIST.json"
        
        # Since stations are sorted alphabetically, we need to find by name instead of index
        # Get the trimmed station name from the selection
        STATION_NAME=$(echo "$SELECTED_TEXT" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        
        # Find the station in the JSON by matching the (trimmed) name
        STATION_DATA=$(jq --arg name "$STATION_NAME" '.[] | select(.name | gsub("^\\s+|\\s+$";"") == $name)' "${LIST_PATH}")
        
        if [ -z "$STATION_DATA" ]; then
            redprint "Could not find station in list."
            menu
            return
        fi
        
        URL_RESOLVED=$(echo "$STATION_DATA" | jq -r '.url_resolved')
        
        if [[ -n $URL_RESOLVED ]] && [[ $URL_RESOLVED != "null" ]]; then
            # Display station info
            clear
            magentaprint "--------- Info Radio: ------------"
            greenprint "NAME: $(echo "$STATION_DATA" | jq -r '.name | gsub("^\\s+|\\s+$";"")')"
            blueprint "TAGS: $(echo "$STATION_DATA" | jq -r '.tags')"
            redprint "COUNTRY: $(echo "$STATION_DATA" | jq -r '.country')"
            yellowprint "VOTES: $(echo "$STATION_DATA" | jq -r '.votes')"
            magentaprint "CODEC: $(echo "$STATION_DATA" | jq -r '.codec')"
            cyanprint "BITRATE: $(echo "$STATION_DATA" | jq -r '.bitrate')"
            magentaprint "----------------------------------"
            echo
            
            _play "$URL_RESOLVED" "$STATION_DATA" "$LIST" || menu
        else
            echo "url_resolved can't be found. Exiting ..."
            exit 1
        fi
    else
        redprint "No such a list."
        exit
    fi
}
