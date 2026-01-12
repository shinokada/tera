#!/usr/bin/env bash

# # opens a favorite list and select number to play a station
# tera play
fn_play() {
    # check if a list is empty
    lists=$(_fav_list)
    # lists=""
    # echo "$lists"
    if [ -z "$lists" ]; then
        redprint "Lists are empty."
        cyanprint "Try $SCRIPT_NAME search"
    fi
    echo
    greenprint "Select a list (use arrow keys, ESC to cancel):"
    echo
    LIST=$(echo "$lists" | tr ' ' '\n' | fzf --prompt="Select a list: " --height=40% --reverse)
    
    # Check if user cancelled
    if [ -z "$LIST" ]; then
        menu
        return
    fi

    # read -rp "Type a list number.   " LIST
    if [ -n "$LIST" ]; then
        echo
        magentaprint "Which station do you want to play?"
        greenprint "Select a station from $LIST list (use arrow keys, ESC to cancel):"
        echo
        STATIONS=$(_station_list "$LIST")
        
        # Use fzf to select station
        SELECTION=$(echo "$STATIONS" | nl | fzf --prompt="Select a station: " --height=40% --reverse)
        
        # Check if user cancelled
        if [ -z "$SELECTION" ]; then
            menu
            return
        fi
        
        # Extract the number from the selection
        ANS=$(echo "$SELECTION" | awk '{print $1}')
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
