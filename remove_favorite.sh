#!/usr/bin/env bash

# Helper script to remove a station from lib/favorite.json
# Usage: ./remove_favorite.sh <index>

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FAVORITE_FILE="$SCRIPT_DIR/lib/favorite.json"

if [ $# -eq 0 ]; then
    echo "Usage: $0 <index>"
    echo ""
    echo "Current favorites:"
    jq -r 'to_entries[] | "\(.key)) \(.value.name)"' "$FAVORITE_FILE"
    exit 1
fi

INDEX=$1

# Check if favorite.json exists
if [ ! -f "$FAVORITE_FILE" ]; then
    echo "Error: No favorites file found"
    exit 1
fi

# Get station name before deleting
STATION_NAME=$(jq -r ".[$INDEX].name" "$FAVORITE_FILE")

if [ "$STATION_NAME" = "null" ]; then
    echo "Error: No station at index $INDEX"
    exit 1
fi

# Remove from favorite.json
TMP_FILE=$(mktemp)
jq "del(.[$INDEX])" "$FAVORITE_FILE" > "$TMP_FILE" && mv "$TMP_FILE" "$FAVORITE_FILE"

echo "✓ Removed '$STATION_NAME' from favorites!"
echo "Total favorite stations: $(jq 'length' "$FAVORITE_FILE")"
