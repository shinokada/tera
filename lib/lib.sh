#!/usr/bin/env bash

### Colors ##
ESC=$(printf '\033')
RESET="${ESC}[0m"
BLACK="${ESC}[30m"
RED="${ESC}[31m"
GREEN="${ESC}[32m"
YELLOW="${ESC}[33m"
BLUE="${ESC}[34m"
MAGENTA="${ESC}[35m"
CYAN="${ESC}[36m"
WHITE="${ESC}[37m"
DEFAULT="${ESC}[39m"

### Color Functions ##

greenprint() {
    printf "${GREEN}%s${RESET}\n" "$1"
}

blueprint() {
    printf "${BLUE}%s${RESET}\n" "$1"
}

redprint() {
    printf "${RED}%s${RESET}\n" "$1"
}

yellowprint() {
    printf "${YELLOW}%s${RESET}\n" "$1"
}

magentaprint() {
    printf "${MAGENTA}%s${RESET}\n" "$1"
}

cyanprint() {
    printf "${CYAN}%s${RESET}\n" "$1"
}

check_cmd() {
    if [[ ! $(command -v "$1") ]]; then
        app=$1
        redprint "It seems like you don't have ${app}."
        redprint "Please install ${app}."
        exit 1
    fi
}

_cleanup_tmp() {
    FILE=$1
    if [[ -f "$FILE" ]]; then
        rm "$FILE"
    fi
}

# shellcheck disable=SC2120
_fav_list() {
    arr=("$FAVORITE_PATH"/*.json)
    list=()
    for f in "${arr[@]}"; do
        if [ $# -ge 1 ] && [ -n "$1" ]; then
            # if any argument is given then show all
            name=$(basename "$f")
            list+=("${name%.*}")
        else
            content_length=$(jq length "$f")
            # echo "$content_length"
            if (("$content_length" > 0)); then
                # otherwise show files with content
                name=$(basename "$f")
                list+=("${name%.*}")
            fi
        fi

    done
    echo "${list[@]}"
}

_graceful_exit() {
    MESSAGE=${1:-"Sorry something went wrong."}
    echo "$MESSAGE"
    exit 1
}

_station_list() {
    jq -r '.[] | .name' <"$FAVORITE_PATH/$1.json"
}

_play() {
    echo
    yellowprint "Press q to quit."
    echo
    mpv "$1" || {
        echo "Not able to play your station."
        return 1
    }
}

_show_favlist() {
    if [ $# -ge 1 ]; then
        LISTS=$(_fav_list "$1")
    else
        LISTS=$(_fav_list)
    fi
    # lists=""
    # echo "$lists"
    if [ -z "$LISTS" ]; then
        redprint "Lists are empty."
        cyanprint "Try $SCRIPT_NAME search"
    fi

    PS3="Enter a number: "
    select LIST in $LISTS; do
        # echo "Selected list: $LIST"
        break
    done
    echo "$LIST"
}

_wget_simple_search() {
    REPLY="$1"
    KEY="$2"
    SEARCH_RESULTS="${TMP_PATH}/radio_searches.json"
    greenprint "Searching ..."
    # echo "Key: $KEY and reply: $REPLY"
    wget --post-data "$KEY=$REPLY" "$SEARCH_URL" -O "$SEARCH_RESULTS" 2>/tmp/tera_error || {
        redprint "Something went wrong. Please see /tmp/tera_error"
        exit
    }
}

_wget_search() {
    REPLY=("$1")
    echo "${REPLY[@]}"
    SEARCH_RESULTS="${TMP_PATH}/radio_searches.json"
    OPTS=
    for TAG in "${REPLY[@]}"; do
        OPTS+="$KEY=$TAG&"
    done
    OPTS="${OPTS%?}"
    greenprint "Searching ..."
    echo "$OPTS"
    wget --post-data "$OPTS" "$SEARCH_URL" -O "$SEARCH_RESULTS" 2>/tmp/tera_error || {
        redprint "Something went wrong. Please see /tmp/tera_error"
        exit
    }
}
