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
    greenprint "Select a list."
    echo
    PS3="Enter a number: "
    select LIST in $(_fav_list); do
        # echo "Selected list: $LIST"
        break
    done

    # read -rp "Type a list number.   " LIST
    if [ -n "$LIST" ]; then
        echo
        magentaprint "Which station do you want to play?"
        greenprint "Select a number to play from $LIST list."
        echo
        STATIONS=$(_station_list "$LIST")
        echo "$STATIONS" | nl
        echo
        printf "Type a number to play. "
        read -r ANS
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
