#!/usr/bin/env bash

fn_delete() {
    TEMP_FILE="$TMP_PATH/radio_delete.json"
    _cleanup_tmp "$TEMP_FILE"
    clear
    cyanprint "$APP_NAME - Delete a Radio Station"
    echo
    touch "$TEMP_FILE"
    greenprint "Select a list number to delete from."
    echo
    echo "0) Main Menu"
    LIST=$(_show_favlist)
    
    # Check if user selected Main Menu during list selection
    if [[ -z $LIST ]] || [[ $LIST == "0" ]]; then
        menu
        return
    fi
    echo
    # echo "LIST: $LIST"
    echo
    echo "     0  Main Menu"
    _station_list "$LIST" | nl
    printf "Type your station number to delete (or 0 for Main Menu): "
    read -r ANS
    # echo "$ANS"
    if [[ -z $ANS ]] || [[ $ANS == "0" ]]; then
        menu
    else
        FAVLIST_PATH="${FAVORITE_PATH}/${LIST}.json"
        jq "del(.[$ANS-1])" <"${FAVLIST_PATH}" >"$TEMP_FILE" && mv "$TEMP_FILE" "$FAVLIST_PATH"
        _station_list "$LIST" | nl
        greenprint "Successfully deleted."
        menu
    fi
}
