#!/usr/bin/env bash

GIST_URL_FILE="$SCRIPT_DOT_DIR/gisturl"

create_gist() {
    clear
    cyanprint "$APP_NAME - Create a Gist"
    echo
    #
    FAV=$(_list_intro)
    # echo "${FAV_ARR[@]}"
    ARR=()
    # add fav list to ARR array
    for file in $FAV; do
        ARR+=("${FAVORITE_PATH}/${file}.json")
    done
    # echo "${ARR[@]}"
    # check GIST_URL exists and has a length greater than zero., if so ask to update else create a Gist
    # if [[ ! -s $GIST_URL_FILE ]]; then
    #     # without mapfile
    #     arr=()
    #     IFS=$'\n' read -ra arr -d '' <"$dotties_file"
    gh gist create -d "Terminal radio favorite lists." "${ARR[@]}" >"$GIST_URL_FILE"
    echo "Created a secret Gist"
    read -r gist_url <"$GIST_URL_FILE"
    # for mac and linux use python3
    python3 -m webbrowser "$gist_url"

    # fi
    gist_menu
}

recover_gist() {
    clear
    cyanprint "$APP_NAME - Recover Favorites from a Gist"
    echo
    greenprint "What is your Gist url?"
    read -r gist_url
    greenprint "Cloning a gist..."
    cd "$FAVORITE_PATH" || {
        redprint "Something went wrong."
        gist_menu
    }
    git clone "$gist_url" || {
        redprint "Your gist url doesn't exist."
        gist_menu
    }
    # find the last from the path
    gist_dir=${gist_url##*/}
    mv "$FAVORITE_PATH"/"$gist_dir"/*.json "$FAVORITE_PATH"
    rm -rf "./$gist_dir"
    greenprint "All your lists are downloaded."
}

gist_menu() {
    clear
    cyanprint "$APP_NAME GIST MENU"
    echo
    
    MENU_OPTIONS="0) Main Menu
1) Create a gist
2) Recover favorites from a gist
3) Exit"
    
    CHOICE=$(echo "$MENU_OPTIONS" | fzf --prompt="Choose an option (arrow keys to navigate): " --height=40% --reverse --no-info)
    
    if [ -z "$CHOICE" ]; then
        menu
        return
    fi
    
    ans=$(echo "$CHOICE" | cut -d')' -f1)
    
    case $ans in
    0)
        menu
        ;;
    1)
        create_gist
        gist_menu
        ;;
    2)
        recover_gist
        gist_menu
        ;;
    3)
        yellowprint "Bye-bye."
        exit 0
        ;;
    *)
        redprint "Wrong option."
        gist_menu
        ;;
    esac
}
