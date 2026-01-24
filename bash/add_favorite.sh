#!/usr/bin/env bash

# Helper script to add a station to lib/favorite.json
# Usage: ./add_favorite.sh <station_json_file>

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FAVORITE_FILE="$SCRIPT_DIR/lib/favorite.json"

if [ $# -eq 0 ]; then
    echo "Usage: $0 <station_json_file>"
    echo ""
    echo "Example:"
    echo "  $0 station.json"
    echo ""
    echo "Or add a station directly from your playlists:"
    echo "  jq '.[0]' ~/.config/tera/favorite/jazz.json | $0 -"
    exit 1
fi

# Check if favorite.json exists
if [ ! -f "$FAVORITE_FILE" ]; then
    echo "[]" > "$FAVORITE_FILE"
fi

# Read station data
if [ "$1" = "-" ]; then
    STATION_DATA=$(cat)
else
    if [ ! -f "$1" ]; then
        echo "Error: File $1 not found"
        exit 1
    fi
    STATION_DATA=$(cat "$1")
fi

# Add to favorite.json
TMP_FILE=$(mktemp)
jq --argjson station "$STATION_DATA" '. += [$station]' "$FAVORITE_FILE" > "$TMP_FILE" && mv "$TMP_FILE" "$FAVORITE_FILE"

STATION_NAME=$(echo "$STATION_DATA" | jq -r '.name')
echo "✓ Added '$STATION_NAME' to favorites!"
echo "Total favorite stations: $(jq 'length' "$FAVORITE_FILE")"
