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
        mpv "$URL_RESOLVED" || {
            echo "Not able to play your station."
            search_menu
        }
    else
        echo "url_resolved can't be found. Exiting ..."
        exit 1
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
    SEARCH_RESULTS="${TMP_PATH}/radio_searches.json"
    echo
    printf "Type a %s to search: " "$KEY"
    read -r REPLY
    echo
    # OPTS=()
    for TAG in "${REPLY[@]}"; do
        OPTS+="$KEY=$TAG&"
    done
    echo "Searching ..."
    wget --post-data "$OPTS" "$SEARCH_URL" -O "$SEARCH_RESULTS" >&/dev/null
    # cat "$SEARCH_RESULTS"
    LENGTH=$(jq length "$SEARCH_RESULTS")

    # check $SEARCH_RESULT has length is more than 0
    if (("$LENGTH" < 1)); then
        echo "No result. Try again."
        search_menu
    fi
    jq -r '.[].name' <"$SEARCH_RESULTS" | nl | fzf
    yellowprint "     0  Return to Search Menu"
    echo
    printf "Select a number. "
    read -r ANS
    # echo "$ANS"
    if [[ "$ANS" == 0 ]]; then
        search_menu
    fi
    # URL_RESOLVED=$(jq -r ".[$ANS-1] |.url_resolved" <"$SEARCH_RESULTS")
    search_submenu "$ANS"
    # rm "$SEARCH_RESULTS"
}

advanced_search() {
    SEARCH_RESULTS="${TMP_PATH}/radio_searches.json"
    # curl -X POST -d 'tag=jazz' -d 'state=queensland' http://all.api.radio-browser.info/json/stations/search
    magentaprint "The query format is -d field=word."
    magentaprint "Field can be one of tag, name, language, country code and state."
    magentaprint "Fields can be combined, for example: -d tag=jazz -d state=queensland"
    magentaprint "Or -d tag=rock -d language=spanish -d countrycode=us"
    magentaprint "Or -d tag=country -d state=IL"
    printf "Write your quiry: "
    read -ra RES
    curl -X POST "${RES[@]}" "$SEARCH_URL" -o "$SEARCH_RESULTS" >&/dev/null
    LENGTH=$(jq length "$SEARCH_RESULTS")

    # check $SEARCH_RESULT has length is more than 0
    if (("$LENGTH" < 1)); then
        echo "No result. Try again."
        search_menu
    fi
    jq -r '.[].name' <"$SEARCH_RESULTS" | nl | fzf
    yellowprint "     0  Return to Search Menu"
    echo
    printf "Select a number. "
    read -r ANS
    if [[ "$ANS" == 0 ]]; then
        search_menu
    fi
    search_submenu "$ANS"
}

search_submenu() {
    echo -ne "
$APP_NAME SEARCH SUBMENU:
$(greenprint '1)') Play
$(greenprint '2)') Save
$(greenprint '3)') Go back to the Search menu
$(greenprint '4)') Go back to the Main menu
$(greenprint '0)') Exit
$(blueprint 'Choose an option:') "
    read -r ans
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
    echo -ne "
$APP_NAME SEARCH MENU:
$(greenprint '1)') Tag
$(greenprint '2)') Name
$(greenprint '3)') Language 
$(greenprint '4)') Country code
$(greenprint '5)') State
$(greenprint '6)') Advanced search
$(greenprint '7)') Main Menu
$(greenprint '0)') Exit
$(blueprint 'Choose an option:') "
    read -r ans
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
