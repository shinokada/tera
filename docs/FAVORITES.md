# Quick Play Favorites Feature

## Overview

TERA now supports **Quick Play Favorites** - a feature that displays your favorite radio stations directly in the main menu for instant access!

## How It Works

When you add stations to `lib/favorite.json`, they automatically appear at the bottom of the main menu under **"Quick Play Favorites"**. You can play them with a single selection!

### Main Menu Display

```
TERA MAIN MENU

1) Play from my list
2) Search radio stations
3) List (Create/Read/Update/Delete)
4) Delete a radio station
5) I feel lucky
6) Gist
0) Exit

--- Quick Play Favorites ---
10) ▶ BBC World Service
11) ▶ Euro Smooth Jazz
12) ▶ Jazz FM
```

## Features

- ✅ **Instant Access** - Play favorites directly from main menu
- ✅ **Visual Indicator** - ▶ symbol shows playable stations
- ✅ **Auto-Display** - Stations appear automatically when added
- ✅ **Limit of 10** - Shows first 10 stations (keeps menu clean)
- ✅ **Station Info** - Displays tags, country, codec before playing

## Managing Favorites

### Method 1: Using Helper Scripts (Easiest)

#### Add a Favorite Station
```bash
# Add from your existing playlists
jq '.[0]' ~/.config/tera/favorite/jazz.json | ./add_favorite.sh -

# Or from a JSON file
./add_favorite.sh station.json
```

#### List All Favorites
```bash
./list_favorites.sh
```

Output:
```
===================================
  Favorite Radio Stations (2)
===================================

0) BBC World Service
   Tags: world,talk,news
   Country: United Kingdom

1) Euro Smooth Jazz
   Tags: lounge,ambient,smooth jazz
   Country: Germany
```

#### Remove a Favorite
```bash
# First, list to see indices
./list_favorites.sh

# Then remove by index
./remove_favorite.sh 0
```

### Method 2: Manual Editing

Edit `lib/favorite.json` directly:

```json
[
  {
    "name": "BBC World Service",
    "url_resolved": "http://stream.live.vc.bbcmedia.co.uk/bbc_world_service",
    "tags": "world,talk,news",
    "country": "The United Kingdom",
    "codec": "MP3",
    "bitrate": 56,
    "votes": 576
  }
]
```

**Required fields:**
- `name` - Station name (displayed in menu)
- `url_resolved` - Stream URL (used for playback)

**Optional but recommended:**
- `tags` - Genre/category tags
- `country` - Country name
- `codec` - Audio codec (MP3, AAC, etc.)
- `bitrate` - Bitrate in kbps
- `votes` - Number of votes

### Method 3: From Within TERA

When you search and find a station you like, save it to a playlist, then add it to favorites:

```bash
# 1. Search for a station in TERA
# 2. Save it to any playlist
# 3. Then use the helper script:
jq '.[0]' ~/.config/tera/favorite/your_list.json | ./add_favorite.sh -
```

## Usage Examples

### Example 1: Adding BBC World Service

```bash
# Create a JSON file with station data
cat > bbc.json << 'EOF'
{
  "name": "BBC World Service",
  "url_resolved": "http://stream.live.vc.bbcmedia.co.uk/bbc_world_service",
  "tags": "world,talk,news",
  "country": "United Kingdom",
  "codec": "MP3",
  "bitrate": 56,
  "votes": 576
}
EOF

# Add it to favorites
./add_favorite.sh bbc.json
```

### Example 2: Adding from Existing Playlist

```bash
# Add the first station from your jazz playlist
jq '.[0]' ~/.config/tera/favorite/jazz.json | ./add_favorite.sh -

# Add the second station
jq '.[1]' ~/.config/tera/favorite/jazz.json | ./add_favorite.sh -
```

### Example 3: Bulk Add Multiple Stations

```bash
# Add first 5 stations from a playlist
for i in {0..4}; do
  jq ".[$i]" ~/.config/tera/favorite/jazz.json | ./add_favorite.sh -
done
```

## File Structure

```
tera/
├── tera                    # Main script (with favorites support)
├── lib/
│   ├── favorite.json      # Your favorite stations (NEW!)
│   ├── sample.json        # Sample stations
│   └── lib.sh            # Added _play_favorite_station()
├── add_favorite.sh        # Helper: Add station
├── remove_favorite.sh     # Helper: Remove station
└── list_favorites.sh      # Helper: List all favorites
```

## Technical Details

### How Favorites are Displayed

1. On main menu load, TERA checks if `lib/favorite.json` exists
2. If it has stations, adds "Quick Play Favorites" section
3. Assigns menu numbers 10-19 to first 10 stations
4. Displays with ▶ symbol for visual clarity

### Menu Numbering

- **1-6**: Standard menu options
- **0**: Exit
- **10-19**: Quick play favorites (dynamic)

### What Happens When You Select a Favorite

1. Extracts station index (e.g., option 10 → index 0)
2. Reads station data from `lib/favorite.json`
3. Displays station information
4. Streams the radio station
5. Returns to main menu when done

## Tips & Best Practices

### 1. Keep It Organized
Only add stations you frequently listen to. The 10-station limit keeps the menu clean.

### 2. Use Descriptive Names
When editing `favorite.json`, use clear station names that help you identify them quickly.

### 3. Test URLs First
Before adding a station, make sure the `url_resolved` works by playing it in TERA first.

### 4. Backup Your Favorites
```bash
cp lib/favorite.json lib/favorite.json.backup
```

### 5. Share Favorites
You can share your `favorite.json` file with friends!

## Troubleshooting

### Favorites Don't Appear in Menu

**Check if file exists and has valid JSON:**
```bash
cat lib/favorite.json
jq '.' lib/favorite.json  # Validates JSON
```

**Check station count:**
```bash
jq 'length' lib/favorite.json
```

### Station Won't Play

**Check if url_resolved exists:**
```bash
jq '.[0].url_resolved' lib/favorite.json
```

**Test URL directly:**
```bash
mpv "$(jq -r '.[0].url_resolved' lib/favorite.json)"
```

### Menu Numbers Overlap

If you have more than 10 favorites, only the first 10 will show (by design). To play others, use "Play from my list" option.

## Advanced: Custom Integration

### Add Favorites from Search Results

Modify the search save function to optionally add to favorites:

```bash
# After saving to a playlist, ask:
read -p "Add to quick play favorites? (y/n) " answer
if [ "$answer" = "y" ]; then
    jq ".[-1]" "$PLAYLIST_FILE" | ./add_favorite.sh -
fi
```

## Benefits

1. **Speed** - Play favorite stations with one click
2. **Convenience** - No need to navigate through playlists
3. **Customization** - Your most-loved stations front and center
4. **Flexibility** - Easy to add/remove via scripts or manual editing

## Future Enhancements

Possible improvements:
- [ ] Add "Manage Favorites" menu option in TERA
- [ ] Sort favorites by name/country/genre
- [ ] Allow more than 10 favorites with pagination
- [ ] Import favorites from URLs or m3u playlists
- [ ] Star rating system

## Questions?

The favorites feature is designed to be simple and powerful. If you have questions or suggestions, feel free to modify the code to suit your needs!

Happy listening! 🎵📻
