#!/usr/bin/env bash

fn_delete() {
    TEMP_FILE="$TMP_PATH/radio_delete.json"
    _cleanup_tmp "$TEMP_FILE"
    echo
    touch "$TEMP_FILE"
    greenprint "Select a list number to delete from."
    echo
    echo "0) CANCEL"
    LIST=$(_show_favlist)
    echo
    # echo "LIST: $LIST"
    if [[ -z $LIST ]]; then
        menu
    fi
    echo "     0  CANCEL"
    _station_list "$LIST" | nl
    printf "Type your station number to delete. "
    read -r ANS
    # echo "$ANS"
    if [[ -z $ANS ]]; then
        menu
    else
        FAVLIST_PATH="${FAVORITE_PATH}/${LIST}.json"
        jq "del(.[$ANS-1])" <"${FAVLIST_PATH}" >"$TEMP_FILE" && mv "$TEMP_FILE" "$FAVLIST_PATH"
        _station_list "$LIST" | nl
        greenprint "Successfully deleted."
        menu
    fi
}
