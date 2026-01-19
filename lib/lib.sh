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
            # Use jq -e to check if file has content (more robust)
            if jq -e 'length > 0' "$f" >/dev/null 2>&1; then
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
    # Trim whitespace and sort alphabetically (case-insensitive)
    # Use .name? | strings to safely handle missing or non-string values
    jq -r '.[] | .name? | strings | gsub("^\\s+|\\s+$";"")' <"$FAVORITE_PATH/$1.json" | sort -f
}

_play() {
    URL=$1
    STATION_DATA=$2
    LIST_NAME=$3
    
    echo
    yellowprint "Press q to quit."
    echo
    mpv "$URL" || {
        echo "Not able to play your station."
        return 1
    }
    
    # After mpv exits, ask if user wants to save to favorites
    if [ -n "$STATION_DATA" ]; then
        _prompt_save_to_favorites "$STATION_DATA" "$LIST_NAME"
    fi
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

_info_select_radio() {
  NAMEINFO=$(cat $SEARCH_RESULTS 2>/dev/null | jq -r ".[$ANS-1]" | grep "name\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  TAGSINFO=$(cat $SEARCH_RESULTS 2>/dev/null | jq -r ".[$ANS-1]" | grep "tags\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  COUNTRYINFO=$(cat $SEARCH_RESULTS 2>/dev/null | jq -r ".[$ANS-1]" | grep "country\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  VOTESINFO=$(cat $SEARCH_RESULTS 2>/dev/null | jq -r ".[$ANS-1]" | grep "votes\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  CODECINFO=$(cat $SEARCH_RESULTS 2>/dev/null | jq -r ".[$ANS-1]" | grep "codec\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  BITRATEINFO=$(cat $SEARCH_RESULTS 2>/dev/null | jq -r ".[$ANS-1]" | grep "bitrate\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  echo
  magentaprint "--------- Info Radio: ------------"
  greenprint "NAME: $NAMEINFO"
  blueprint "TAGS: $TAGSINFO"
  redprint "COUNTRY: $COUNTRYINFO"
  yellowprint "VOTES: $VOTESINFO"
  magentaprint "CODEC: $CODECINFO"
  cyanprint "BITRATE: $BITRATEINFO"
  magentaprint "---------------------------------- "
  echo
}

_info_select_radio_play() {
  NAMEINFO=$(cat $LIST_PATH 2>/dev/null | jq -r ".[$ANS-1]" | grep "name\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  TAGSINFO=$(cat $LIST_PATH 2>/dev/null | jq -r ".[$ANS-1]" | grep "tags\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  COUNTRYINFO=$(cat $LIST_PATH 2>/dev/null | jq -r ".[$ANS-1]" | grep "country\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  VOTESINFO=$(cat $LIST_PATH 2>/dev/null | jq -r ".[$ANS-1]" | grep "votes\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  CODECINFO=$(cat $LIST_PATH 2>/dev/null | jq -r ".[$ANS-1]" | grep "codec\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  BITRATEINFO=$(cat $LIST_PATH 2>/dev/null | jq -r ".[$ANS-1]" | grep "bitrate\":" | awk -F': ' '{print $2}' | sed 's/"//g' | sed 's/+/ /g' | sed 's/,//g')
  clear
  magentaprint "--------- Info Radio: ------------"
  greenprint "NAME: $NAMEINFO"
  blueprint "TAGS: $TAGSINFO"
  redprint "COUNTRY: $COUNTRYINFO"
  yellowprint "VOTES: $VOTESINFO"
  magentaprint "CODEC: $CODECINFO"
  cyanprint "BITRATE: $BITRATEINFO"
  magentaprint "---------------------------------- "
  echo
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
    # Clear the "Searching..." message
    echo -ne "\r\033[K"
}

_wget_search() {
    read -rp 'Write your query ' -a REPLY
    echo "${REPLY[@]}"
    SEARCH_RESULTS="${TMP_PATH}/radio_searches.json"
    greenprint "Searching ..."
    curl -X POST "${REPLY[@]}" "$SEARCH_URL" > "$SEARCH_RESULTS" 2>/tmp/tera_error || {
        redprint "Something went wrong. Please see /tmp/tera_error"
        exit
    }
    # Clear the "Searching..." message
    echo -ne "\r\033[K"
}

_play_favorite_station() {
    STATION_INDEX=$1
    # Use the user's My Favorites list from config directory (My-favorites.json)
    # This is the same file users save to when they select "My Favorites"
    FAVORITE_STATIONS_FILE="${FAVORITE_PATH}/My-favorites.json"
    
    if [ ! -f "$FAVORITE_STATIONS_FILE" ]; then
        redprint "Favorite stations file not found."
        return 1
    fi
    
    # Get station data
    STATION_DATA=$(jq -r ".[${STATION_INDEX}]" "$FAVORITE_STATIONS_FILE" 2>/dev/null)
    URL_RESOLVED=$(echo "$STATION_DATA" | jq -r '.url_resolved')
    STATION_NAME=$(echo "$STATION_DATA" | jq -r '.name')
    
    if [ -z "$URL_RESOLVED" ] || [ "$URL_RESOLVED" = "null" ]; then
        redprint "Could not find station URL."
        return 1
    fi
    
    # Display station info
    clear
    cyanprint "$APP_NAME - Playing Favorite Station"
    echo
    _info_favorite_station "$STATION_INDEX" "$FAVORITE_STATIONS_FILE"
    
    # Play the station - no need to prompt for favorites since it's already in favorites
    # Pass empty string for station data and list name to skip the prompt
    _play "$URL_RESOLVED" "" ""
}

_info_favorite_station() {
    STATION_INDEX=$1
    FAV_FILE=$2
    
    NAMEINFO=$(jq -r ".[${STATION_INDEX}].name" "$FAV_FILE" 2>/dev/null)
    TAGSINFO=$(jq -r ".[${STATION_INDEX}].tags" "$FAV_FILE" 2>/dev/null)
    COUNTRYINFO=$(jq -r ".[${STATION_INDEX}].country" "$FAV_FILE" 2>/dev/null)
    VOTESINFO=$(jq -r ".[${STATION_INDEX}].votes" "$FAV_FILE" 2>/dev/null)
    CODECINFO=$(jq -r ".[${STATION_INDEX}].codec" "$FAV_FILE" 2>/dev/null)
    BITRATEINFO=$(jq -r ".[${STATION_INDEX}].bitrate" "$FAV_FILE" 2>/dev/null)
    
    magentaprint "--------- Info Radio: ------------"
    greenprint "NAME: $NAMEINFO"
    blueprint "TAGS: $TAGSINFO"
    redprint "COUNTRY: $COUNTRYINFO"
    yellowprint "VOTES: $VOTESINFO"
    magentaprint "CODEC: $CODECINFO"
    cyanprint "BITRATE: $BITRATEINFO"
    magentaprint "----------------------------------"
    echo
}

# Prompt user to save station to Quick Play Favorites after playing
_prompt_save_to_favorites() {
    STATION_DATA=$1
    LIST_NAME=$2
    
    clear
    cyanprint "Did you enjoy this station?"
    echo
    
    STATION_NAME=$(echo "$STATION_DATA" | jq -r '.name | gsub("^\\s+|\\s+$";"")')
    greenprint "Station: $STATION_NAME"
    if [ -n "$LIST_NAME" ]; then
        blueprint "From list: $LIST_NAME"
    fi
    echo
    
    OPTIONS="1) ⭐ Add to Quick Play Favorites
2) Return to Main Menu"
    
    CHOICE=$(echo "$OPTIONS" | fzf --prompt="Choose an option: " --height=40% --reverse --no-info)
    
    # User cancelled or no choice
    if [ -z "$CHOICE" ]; then
        return 0
    fi
    
    ANS=$(echo "$CHOICE" | cut -d')' -f1)
    
    case $ANS in
    1)
        _add_to_quick_favorites "$STATION_DATA"
        ;;
    2)
        return 0
        ;;
    *)
        return 0
        ;;
    esac
}

# Add station to Quick Play Favorites (My-favorites.json)
_add_to_quick_favorites() {
    STATION_DATA=$1
    FAVORITES_FILE="${FAVORITE_PATH}/My-favorites.json"
    
    # Initialize file if it doesn't exist
    if [ ! -f "$FAVORITES_FILE" ]; then
        echo "[]" > "$FAVORITES_FILE"
    fi
    
    # Get station UUID for duplicate check
    STATION_UUID=$(echo "$STATION_DATA" | jq -r '.stationuuid')
    
    # Check if station already exists
    EXISTS=$(jq --arg uuid "$STATION_UUID" 'any(.[]; .stationuuid == $uuid)' "$FAVORITES_FILE")
    
    if [ "$EXISTS" = "true" ]; then
        yellowprint "⭐ Station is already in Quick Play Favorites!"
        echo
        yellowprint "You can access it from the Main Menu."
        sleep 2
        return 0
    fi
    
    # Add station to favorites
    TEMP_FILE="${FAVORITES_FILE}.tmp"
    jq --argjson station "$STATION_DATA" '. += [$station]' "$FAVORITES_FILE" > "$TEMP_FILE"
    mv "$TEMP_FILE" "$FAVORITES_FILE"
    
    greenprint "✓ Added to Quick Play Favorites!"
    echo
    yellowprint "You can now access this station from the Main Menu."
    sleep 2
}
