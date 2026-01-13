#!/usr/bin/env bash

_list_intro() {
    lists=$(_fav_list)
    if [ -z "$lists" ]; then
        redprint "Lists are empty"
        cyanprint "Try $SCRIPT_NAME search"
    fi
    for list in $(_fav_list all); do
        echo "$list"
    done
}

create_list() {
    echo
    greenprint "My lists: "
    _list_intro
    echo
    printf "Type a new list name: "
    read -r NEW_LIST
    echo
    # replace spaces with - in $NEW_LIST
    NEW=$NEW_LIST
    NAME="${NEW// /-}"
    touch "$FAVORITE_PATH/$NAME.json"
    echo "[]" >"$FAVORITE_PATH/$NAME.json"
    greenprint "$NAME is created."
    list_menu
}

delete_list() {
    echo
    greenprint "My lists: "
    _list_intro
    echo
    printf "Type a list name to delete: "
    read -r LIST
    echo
    rm "$FAVORITE_PATH/$LIST.json" 2>/dev/null || {
        redprint "$LIST doesn't exist. Try it again."
        delete_list
    }
    greenprint "$LIST is deleted"
    echo
    greenprint "My lists: "
    _list_intro
    echo
    list_menu
}

show_lists() {
    echo
    greenprint "My lists: "
    _list_intro
    echo
}

edit_list() {
    echo
    greenprint "My lists: "
    _list_intro
    echo
    printf "Type a list name to edit: "
    read -r LIST
    yellowprint "Old name: $LIST"
    if [ ! -f "$FAVORITE_PATH/$LIST.json" ]; then
        redprint "$LIST doesn't exist. Try again."
        edit_list
    fi
    printf "Type a new name: "
    read -r NEW
    NAME="${NEW// /-}"
    cyanprint "New name: $NAME"
    mv "$FAVORITE_PATH/$LIST.json" "$FAVORITE_PATH/$NAME.json" &>/dev/null || {
        redprint "Something went wrong. Try it again."
        list_menu
    }
    greenprint "Updated the list name."
    show_lists
}

list_menu() {
    clear
    cyanprint "$APP_NAME LIST MENU"
    echo
    
    MENU_OPTIONS="0) Main Menu
1) Create a list
2) Delete a list
3) Edit a list name
4) Show all list names
5) Exit"
    
    CHOICE=$(echo "$MENU_OPTIONS" | fzf --prompt="Choose an option (arrow keys to navigate): " --height=40% --reverse --no-info)
    
    if [ -z "$CHOICE" ]; then
        menu
        return
    fi
    
    ans=$(echo "$CHOICE" | cut -d')' -f1)
    
    case $ans in
    0)
        echo "Go back to the main menu"
        menu
        ;;
    1)
        create_list
        list_menu
        ;;
    2)
        delete_list
        list_menu
        ;;
    3)
        edit_list
        list_menu
        ;;
    4)
        show_lists
        list_menu
        ;;
    5)
        yellowprint "Bye-bye."
        exit 0
        ;;
    *)
        redprint "Wrong option."
        list_menu
        ;;
    esac
}
