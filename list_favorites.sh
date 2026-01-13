#!/usr/bin/env bash

# Helper script to list all stations in lib/favorite.json
# Usage: ./list_favorites.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FAVORITE_FILE="$SCRIPT_DIR/lib/favorite.json"

if [ ! -f "$FAVORITE_FILE" ]; then
    echo "No favorites file found"
    exit 1
fi

COUNT=$(jq 'length' "$FAVORITE_FILE")

if [ "$COUNT" -eq 0 ]; then
    echo "No favorite stations yet"
    exit 0
fi

echo "==================================="
echo "  Favorite Radio Stations ($COUNT)"
echo "==================================="
echo ""

jq -r 'to_entries[] | "\(.key)) \(.value.name)\n   Tags: \(.value.tags)\n   Country: \(.value.country)\n"' "$FAVORITE_FILE"
