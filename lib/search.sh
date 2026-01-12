#!/usr/bin/env bash

_save_station_to_list() {
    TEMP_FILE="${TMP_PATH}/radio_favorite.json"
    TEMP_FILE2="${TMP_PATH}/radio_favorite2.json"
    _cleanup_tmp "$TEMP_FILE"
    _cleanup_tmp "$TEMP_FILE2"
    echo
    ANS=$1
    # echo $ANS
    greenprint "Select the list to save to: "
    # echo saving
    LIST_NAME=$(_show_favlist all)
    echo "$LIST_NAME"
    FAVORITE_FULL="${FAVORITE_PATH}/${LIST_NAME}.json"
    # echo "$FAVORITE_FULL"
    # get item from "$SEARCH_RESULTS" using $ANS

    jq ".[$ANS-1]" <"$SEARCH_RESULTS" >"$TEMP_FILE"
    # add the item to the fav list
    # jq '. += [input]' "$FAVORITE_FULL" "$TEMP_FILE"
    jq '. += [input]' "$FAVORITE_FULL" "$TEMP_FILE" >"$TEMP_FILE2" && mv "$TEMP_FILE2" "$FAVORITE_FULL"
    # which list?
    echo
    greenprint "Successfully saved the station to your $LIST_NAME list."
}

_search_play() {
    # FILE="/tmp/radio_listening.json"
    TEMP_FILE="${TMP_PATH}/radio_favorite.json"
    # TEMP_FILE2="${TMP_PATH}/radio_favorite2.json"
    _cleanup_tmp "$TEMP_FILE"
    # echo "$1" # this is a list number in "$SEARCH_RESULTS"
    ANS=$1
    jq -r ".[$ANS-1]" <"$SEARCH_RESULTS" >"$TEMP_FILE"
    URL_RESOLVED=$(jq -r ".[$ANS-1] |.url_resolved" <"$SEARCH_RESULTS")
    if [ -n "$URL_RESOLVED" ]; then
        _info_select_radio "$ANS"
        mpv "$URL_RESOLVED" || {
            echo "Not able to play your station."
            search_menu
        }
    else
        redprint "url_resolved can't be found. Returning to search menu..."
        search_menu
        return
    fi
    echo
    printf "Do you want to save this station? (yes/y/no/n) "
    read -r RES
    USER_ANS=$(echo "$RES" | cut -c 1-1 | tr "[:lower:]" "[:upper:]")
    if [ "$USER_ANS" = "Y" ]; then
        _save_station_to_list "$ANS"
    fi
}

search_by() {
    KEY=$1
    echo
    printf "Type a %s to search: " "$KEY"
    read -r REPLY
    echo
    # OPTS=()
    _wget_simple_search "$REPLY" "$KEY"
    # SEARCH_RESULTS="${TMP_PATH}/radio_searches.json"
    # for TAG in "${REPLY[@]}"; do
    #     OPTS+="$KEY=$TAG&"
    # done
    # greenprint "Searching ..."
    # wget --post-data "$OPTS" "$SEARCH_URL" -O "$SEARCH_RESULTS" 2>/tmp/tera_error || {
    #     redprint "Something went wrong. Please see /tmp/tera_error"
    #     exit
    # }
    # cat "$SEARCH_RESULTS"
    LENGTH=$(jq length "$SEARCH_RESULTS")

    # check $SEARCH_RESULT has length is more than 0
    if (("$LENGTH" < 1)); then
        yellowprint "No result. Try again."
        search_menu
    fi
    # Use fzf to interactively select a station
    SELECTION=$(jq -r '.[].name' <"$SEARCH_RESULTS" | nl | fzf --prompt="Select a station (or ESC to return): " --height=40% --reverse)
    
    # Check if user cancelled (ESC)
    if [ -z "$SELECTION" ]; then
        search_menu
        return
    fi
    
    # Extract the number from the selection
    ANS=$(echo "$SELECTION" | awk '{print $1}')
    # URL_RESOLVED=$(jq -r ".[$ANS-1] |.url_resolved" <"$SEARCH_RESULTS")
    _info_select_radio "$ANS"
    search_submenu "$ANS"
    # rm "$SEARCH_RESULTS"
}

advanced_search() {
    SEARCH_RESULTS="${TMP_PATH}/radio_searches.json"

    magentaprint "The query format is -d field=word."
    magentaprint "Field can be one of tag, name, language, country code and state."
    magentaprint "Fields can be combined, for example: -d tag=jazz -d state=queensland"
    magentaprint "Or -d tag=rock -d language=spanish -d countrycode=us"
    magentaprint "Or -d tag=jazz -d codec=ogg -d bitrateMin=128000"
    _wget_search

    LENGTH=$(jq length "$SEARCH_RESULTS")

    # check $SEARCH_RESULT has length is more than 0
    if (("$LENGTH" < 1)); then
        echo "No result. Try again."
        search_menu
    fi
    # Use fzf to interactively select a station
    SELECTION=$(jq -r '.[].name' <"$SEARCH_RESULTS" | nl | fzf --prompt="Select a station (or ESC to return): " --height=40% --reverse)
    
    # Check if user cancelled (ESC)
    if [ -z "$SELECTION" ]; then
        search_menu
        return
    fi
    
    # Extract the number from the selection
    ANS=$(echo "$SELECTION" | awk '{print $1}')
    _info_select_radio "$ANS"
    search_submenu "$ANS"
}

search_submenu() {
    clear
    cyanprint "$APP_NAME SEARCH SUBMENU"
    echo
    
    MENU_OPTIONS="1) Play
2) Save
3) Go back to the Search menu
4) Go back to the Main menu
0) Exit"
    
    CHOICE=$(echo "$MENU_OPTIONS" | fzf --prompt="Choose an option (arrow keys to navigate): " --height=40% --reverse --no-info)
    
    if [ -z "$CHOICE" ]; then
        search_menu
        return
    fi
    
    ans=$(echo "$CHOICE" | cut -d')' -f1)
    
    case $ans in
    1)
        _search_play "$ANS"
        search_menu
        ;;
    2)
        _save_station_to_list "$ANS"
        search_menu
        ;;
    3)
        echo "Go back to the Search menu"
        search_menu
        ;;
    4)
        echo "Go back to the Main menu"
        menu
        ;;
    0)
        yellowprint "Bye-bye."
        exit 0
        ;;
    *)
        redprint 'Wrong option.'
        search_menu
        ;;
    esac
}

search_menu() {
    _cleanup_tmp "${TMP_PATH}/radio_searches.json"
    clear
    cyanprint "$APP_NAME SEARCH MENU"
    echo
    
    MENU_OPTIONS="1) Tag
2) Name
3) Language
4) Country code
5) State
6) Advanced search
7) Main Menu
0) Exit"
    
    CHOICE=$(echo "$MENU_OPTIONS" | fzf --prompt="Choose an option (arrow keys to navigate): " --height=40% --reverse --no-info)
    
    if [ -z "$CHOICE" ]; then
        menu
        return
    fi
    
    ans=$(echo "$CHOICE" | cut -d')' -f1)
    
    case $ans in
    1)
        search_by tag
        search_menu
        ;;
    2)
        search_by name
        search_menu
        ;;
    3)
        echo "Search by language"
        search_by language
        search_menu
        ;;
    4)
        search_by countrycode
        search_menu
        ;;
    5)
        search_by state
        search_menu
        ;;
    6)
        advanced_search
        search_menu
        ;;
    7)
        menu
        ;;
    0)
        yellowprint "Bye-bye."
        exit 0
        ;;
    *)
        redprint "Wrong option."
        search_menu
        ;;
    esac
}
