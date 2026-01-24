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
    clear
    cyanprint "$APP_NAME - Create New List"
    echo
    greenprint "My lists: "
    _list_intro
    echo
    yellowprint "Type '0' to go back, '00' for main menu"
    printf "Type a new list name: "
    read -r NEW_LIST
    echo
    
    # Check for navigation commands
    case "$NEW_LIST" in
        "0"|"back")
            list_menu
            return
            ;;
        "00"|"main")
            menu
            return
            ;;
        "")
            redprint "List name cannot be empty."
            sleep 1
            create_list
            return
            ;;
    esac
    
    # replace spaces with - in $NEW_LIST
    NEW=$NEW_LIST
    NAME="${NEW// /-}"
    
    # Check if list already exists
    if [ -f "$FAVORITE_PATH/$NAME.json" ]; then
        redprint "List '$NAME' already exists!"
        sleep 1
        create_list
        return
    fi
    
    touch "$FAVORITE_PATH/$NAME.json"
    echo "[]" >"$FAVORITE_PATH/$NAME.json"
    greenprint "$NAME is created."
    sleep 1
    list_menu
}

delete_list() {
    clear
    cyanprint "$APP_NAME - Delete List"
    echo
    greenprint "My lists: "
    _list_intro
    echo
    yellowprint "Type '0' to go back, '00' for main menu"
    printf "Type a list name to delete: "
    read -r LIST
    echo
    
    # Check for navigation commands
    case "$LIST" in
        "0"|"back")
            list_menu
            return
            ;;
        "00"|"main")
            menu
            return
            ;;
        "")
            redprint "Please enter a list name."
            sleep 1
            delete_list
            return
            ;;
    esac
    
    # Check if it's a special file (My-favorites)
    if [ "$LIST" = "My-favorites" ]; then
        redprint "Cannot delete My-favorites list!"
        sleep 1
        delete_list
        return
    fi
    
    rm "$FAVORITE_PATH/$LIST.json" 2>/dev/null || {
        redprint "$LIST doesn't exist. Try it again."
        sleep 1
        delete_list
        return
    }
    greenprint "$LIST is deleted"
    sleep 1
    list_menu
}

show_lists() {
    clear
    cyanprint "$APP_NAME - All Lists"
    echo
    greenprint "My lists: "
    _list_intro
    echo
    yellowprint "Press Enter to continue..."
    read -r
}

edit_list() {
    clear
    cyanprint "$APP_NAME - Edit List Name"
    echo
    greenprint "My lists: "
    _list_intro
    echo
    yellowprint "Type '0' to go back, '00' for main menu"
    printf "Type a list name to edit: "
    read -r LIST
    echo
    
    # Check for navigation commands
    case "$LIST" in
        "0"|"back")
            list_menu
            return
            ;;
        "00"|"main")
            menu
            return
            ;;
        "")
            redprint "Please enter a list name."
            sleep 1
            edit_list
            return
            ;;
    esac
    
    yellowprint "Old name: $LIST"
    
    if [ ! -f "$FAVORITE_PATH/$LIST.json" ]; then
        redprint "$LIST doesn't exist. Try again."
        sleep 1
        edit_list
        return
    fi
    
    # Check if it's a special file (My-favorites)
    if [ "$LIST" = "My-favorites" ]; then
        redprint "Cannot rename My-favorites list!"
        sleep 1
        edit_list
        return
    fi
    
    yellowprint "Type '0' to go back, '00' for main menu"
    printf "Type a new name: "
    read -r NEW
    echo
    
    # Check for navigation commands
    case "$NEW" in
        "0"|"back")
            list_menu
            return
            ;;
        "00"|"main")
            menu
            return
            ;;
        "")
            redprint "New name cannot be empty."
            sleep 1
            edit_list
            return
            ;;
    esac
    
    NAME="${NEW// /-}"
    
    # Check if new name already exists
    if [ -f "$FAVORITE_PATH/$NAME.json" ]; then
        redprint "List '$NAME' already exists!"
        sleep 1
        edit_list
        return
    fi
    
    cyanprint "New name: $NAME"
    mv "$FAVORITE_PATH/$LIST.json" "$FAVORITE_PATH/$NAME.json" &>/dev/null || {
        redprint "Something went wrong. Try it again."
        sleep 1
        list_menu
        return
    }
    greenprint "Updated the list name."
    sleep 1
    list_menu
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
        menu
        ;;
    1)
        create_list
        ;;
    2)
        delete_list
        ;;
    3)
        edit_list
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
