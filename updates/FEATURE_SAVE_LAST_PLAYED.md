# Feature Proposal: Save Last Played Station to Quick Play Favorites

## Current Workflow Problem

**User Journey:**
1. User plays station from "Play from My List"
2. Station plays (mpv running)
3. User presses `q` to quit
4. Returns to MAIN MENU
5. **Problem:** User liked the station but has no quick way to save it to Quick Play Favorites (My-favorites.json)

**Current workaround:**
- User must remember station name
- Navigate to List menu
- Find the original list
- Find the station again
- Move it to My-favorites
- **This is tedious!**

---

## Proposed Solution

### Option 1: Post-Play Prompt (RECOMMENDED)

**Flow:**
```
[Station playing in mpv]
↓
User presses 'q' to quit
↓
mpv exits, returns to TERA
↓
┌─────────────────────────────────────────┐
│ Did you enjoy this station?             │
│                                          │
│ 1) ⭐ Add to Quick Play Favorites       │
│ 2) Return to Main Menu                  │
└─────────────────────────────────────────┘
↓
If user selects 1:
  - Check if station already in My-favorites.json
  - If not, add it
  - Show confirmation: "✓ Added to Quick Play Favorites!"
  - Return to Main Menu
```

**Implementation location:** Modify `_play()` function in `lib/lib.sh`

**Pros:**
- ✅ Immediate feedback - "I just liked this!"
- ✅ Contextual - right after listening
- ✅ Simple UX - one or two keystrokes
- ✅ Non-intrusive - can skip by selecting option 2

**Cons:**
- ⚠️ Adds one extra step after every song (but can be quick)

---

### Option 2: Save During Play

**Flow:**
```
[Station playing in mpv]
↓
Press 'f' (or another key) to favorite while playing
↓
Shows notification: "⭐ Added to Quick Play Favorites!"
↓
Continue playing
```

**Implementation:** Would require modifying mpv key bindings

**Pros:**
- ✅ Can favorite while still listening
- ✅ Very quick

**Cons:**
- ❌ Requires mpv configuration
- ❌ More complex implementation
- ❌ User might not discover the feature
- ❌ Conflicts with mpv's default keybindings

---

### Option 3: Main Menu Option (LEAST RECOMMENDED)

Add a new menu option: "7) Add last played station to favorites"

**Pros:**
- ✅ Simple to implement

**Cons:**
- ❌ User must remember what they last played
- ❌ Not contextual
- ❌ Extra menu clutter
- ❌ Doesn't work if user played multiple stations

---

## Recommended Implementation: Option 1

### Where to Add

**Primary changes in `lib/lib.sh`:**

```bash
# Global variable to track last played station
LAST_PLAYED_STATION=""
LAST_PLAYED_LIST=""

_play() {
    URL=$1
    STATION_DATA=$2  # Pass station data as second parameter
    LIST_NAME=$3     # Pass list name as third parameter
    
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
    esac
}

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
        sleep 1
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
```

**Changes in `lib/play.sh`:**

```bash
# In fn_play(), when calling _play:
_play "$URL_RESOLVED" "$STATION_DATA" "$LIST" || menu
```

**Changes in other play locations:**
- `lib/search.sh` - When playing from search results
- `lib/lucky.sh` - When playing from I Feel Lucky
- Main menu quick play favorites

---

## Implementation Checklist

### Step 1: Update lib/lib.sh
- [ ] Add `_prompt_save_to_favorites()` function
- [ ] Add `_add_to_quick_favorites()` function
- [ ] Modify `_play()` to accept station data and call prompt after playback

### Step 2: Update lib/play.sh
- [ ] Modify `fn_play()` to pass `STATION_DATA` and `LIST` to `_play()`

### Step 3: Update lib/search.sh
- [ ] Modify search play calls to pass station data to `_play()`

### Step 4: Update lib/lucky.sh
- [ ] Modify lucky play calls to pass station data to `_play()`

### Step 5: Update main tera script
- [ ] Modify `_play_favorite_station()` to pass station data

### Step 6: Testing
- [ ] Test from "Play from My List"
- [ ] Test from "Search"
- [ ] Test from "I Feel Lucky"
- [ ] Test from Quick Play menu items
- [ ] Test duplicate detection
- [ ] Test when My-favorites.json doesn't exist
- [ ] Test cancellation (pressing ESC or selecting option 2)

---

## User Experience

### Before
```
User plays station → Likes it → Presses q → Back to menu
→ User thinks "I wish I could save that..."
→ Has to navigate back through menus to find and save it
```

### After
```
User plays station → Likes it → Presses q
→ "Did you enjoy this station?"
→ Press 1 (or arrow down + Enter)
→ "✓ Added to Quick Play Favorites!"
→ Back to menu
→ Station now appears in main menu for instant access
```

---

## Alternative: Simplified Version (Quick Implementation)

If you want something simpler to start with:

```bash
_play() {
    URL=$1
    STATION_DATA=$2
    
    echo
    yellowprint "Press q to quit."
    echo
    mpv "$URL" || return 1
    
    # Simple prompt
    if [ -n "$STATION_DATA" ]; then
        echo
        printf "Add to Quick Play Favorites? (y/n): "
        read -r ANSWER
        if [ "$ANSWER" = "y" ] || [ "$ANSWER" = "Y" ]; then
            _add_to_quick_favorites "$STATION_DATA"
        fi
    fi
}
```

**Pros:**
- Much simpler
- No fzf needed for this prompt
- Quick yes/no

**Cons:**
- Not as pretty
- No visual feedback before answering
- Easy to accidentally skip

---

## Recommendation

**I recommend Option 1 (Post-Play Prompt) with the fzf interface** because:

1. ✅ **Natural workflow** - Right after listening, while it's fresh
2. ✅ **Discoverable** - User sees the option every time
3. ✅ **Consistent UX** - Uses same fzf interface as rest of TERA
4. ✅ **Non-intrusive** - Can quickly select option 2 to skip
5. ✅ **Complete** - Handles duplicates, missing files, etc.
6. ✅ **Flexible** - Works from any play context (lists, search, lucky)

**Next steps:**
1. Would you like me to implement this?
2. Any modifications to the proposed UX?
3. Should we add any additional options (like "Add to different list")?
