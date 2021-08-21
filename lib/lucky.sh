#!/usr/bin/env bash

fn_lucky() {
    _cleanup_tmp "${TMP_PATH}/radio_searches.json"
    echo
    magentaprint "Type a genre of music, rock, classical, jazz, pop, country, hip, heavy, blues, soul."
    magentaprint "Or type a keyword, like meditation, relax, mozart, Beatles etc."
    cyanprint "Use only one word."
    echo
    # ask a tag word
    printf "Genre/keyword: "
    read -r REPLY
    # find all stations with a key word = tag
    _wget_simple_search "$REPLY" "tag"
    # find the list length
    LENGTH=$(jq length "$SEARCH_RESULTS")
    # random number
    if (("$LENGTH" > 0)); then
        RAN_NUM=$((1 + RANDOM % LENGTH))
        # echo "$RAN_NUM"
        _search_play "$RAN_NUM"
    else
        redprint "No results. Please try it again."
        fn_lucky
    fi
}
