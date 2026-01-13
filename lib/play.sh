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
        # echo "$ANS"
        # get list path
        # FAV_FULL=_station_list
        # echo "LIST: $LIST"
        # echo "$FAVORITE_PATH/$LIST.json"
        LIST_PATH="$FAVORITE_PATH/$LIST.json"
        # echo "${LIST[$ANS]}"
        # find the $ANS line e.g. line 2
        URL_RESOLVED=$(jq -r ".[$ANS-1] |.url_resolved" <"${LIST_PATH}")
        # echo "url_resolved: $URL_RESOLVED"
        if [[ -n $URL_RESOLVED ]]; then
            _info_select_radio_play "$ANS" "${LIST_PATH}"
            _play "$URL_RESOLVED" || menu
        else
            echo "url_resolved can't be found. Exiting ..."
            exit 1
        fi
    else
        redprint "No such a list."
        exit
    fi
}
