# Quick Start: Adding Your First Favorite Station

## Step 1: Find a Station You Love

Start TERA and search for a station:
```bash
./tera
# Select: 2) Search radio stations
# Search by tag: jazz
# Play and find one you like
# Save it to any playlist
```

## Step 2: Add It to Favorites

### Option A: From Your Playlist
```bash
# Make scripts executable (first time only)
chmod +x add_favorite.sh remove_favorite.sh list_favorites.sh

# Add first station from your playlist
jq '.[0]' ~/.config/tera/favorite/YOUR_PLAYLIST.json | ./add_favorite.sh -
```

### Option B: Create a Simple JSON File
```bash
cat > my_station.json << 'EOF'
{
  "name": "My Favorite Jazz Station",
  "url_resolved": "http://streaming.example.com/jazz"
}
EOF

./add_favorite.sh my_station.json
```

### Option C: Use Sample Stations
```bash
# Add BBC World Service from sample
jq '.[0]' lib/sample.json | ./add_favorite.sh -

# Add Euro Smooth Jazz from sample
jq '.[1]' lib/sample.json | ./add_favorite.sh -
```

## Step 3: Enjoy!

Restart TERA and you'll see your favorites in the main menu:

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
```

Just select option 10 or 11 to play instantly!

## Common Commands

```bash
# List your favorites
./list_favorites.sh

# Remove a favorite (after checking list)
./remove_favorite.sh 0

# Add from existing playlist
jq '.[INDEX]' ~/.config/tera/favorite/PLAYLIST.json | ./add_favorite.sh -
```

That's it! You now have quick-play favorites. 🎵

For more details, see [FAVORITES.md](FAVORITES.md)
