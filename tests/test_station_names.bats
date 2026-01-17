#!/usr/bin/env bats

# Tests for station name trimming and alphabetical sorting improvements
# Created: January 17, 2026

setup() {
    # Load the library functions
    export SCRIPT_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)"
    source "$SCRIPT_DIR/lib/lib.sh"
    
    # Create temporary test directory
    export TEST_DIR="$(mktemp -d)"
    export FAVORITE_PATH="$TEST_DIR/favorites"
    mkdir -p "$FAVORITE_PATH"
    
    # Create test JSON file with stations (including ones with whitespace)
    cat > "$FAVORITE_PATH/test-list.json" << 'EOF'
[
  {
    "name": "  BBC Radio 1  ",
    "url_resolved": "http://example.com/bbc1",
    "tags": "pop,music",
    "country": "UK",
    "votes": 100,
    "codec": "MP3",
    "bitrate": 128,
    "stationuuid": "uuid-1"
  },
  {
    "name": "Jazz FM",
    "url_resolved": "http://example.com/jazz",
    "tags": "jazz",
    "country": "UK",
    "votes": 50,
    "codec": "AAC",
    "bitrate": 64,
    "stationuuid": "uuid-2"
  },
  {
    "name": "Classical Music  ",
    "url_resolved": "http://example.com/classical",
    "tags": "classical",
    "country": "US",
    "votes": 75,
    "codec": "MP3",
    "bitrate": 128,
    "stationuuid": "uuid-3"
  },
  {
    "name": "  Smooth Jazz",
    "url_resolved": "http://example.com/smooth",
    "tags": "jazz,smooth",
    "country": "US",
    "votes": 60,
    "codec": "AAC",
    "bitrate": 96,
    "stationuuid": "uuid-4"
  },
  {
    "name": "Rock Station",
    "url_resolved": "http://example.com/rock",
    "tags": "rock",
    "country": "US",
    "votes": 90,
    "codec": "MP3",
    "bitrate": 192,
    "stationuuid": "uuid-5"
  }
]
EOF
}

teardown() {
    # Clean up test directory
    rm -rf "$TEST_DIR"
}

# Test 1: Verify _station_list trims whitespace
@test "station names have whitespace trimmed" {
    result=$(_station_list "test-list")
    
    # Check that no line has leading or trailing spaces
    while IFS= read -r line; do
        # Check for leading whitespace
        if [[ "$line" =~ ^[[:space:]] ]]; then
            echo "Line has leading whitespace: '$line'"
            return 1
        fi
        # Check for trailing whitespace
        if [[ "$line" =~ [[:space:]]$ ]]; then
            echo "Line has trailing whitespace: '$line'"
            return 1
        fi
    done <<< "$result"
    
    return 0
}

# Test 2: Verify _station_list returns alphabetically sorted results
@test "stations are sorted alphabetically (case-insensitive)" {
    result=$(_station_list "test-list")
    
    # Expected order (alphabetical, case-insensitive)
    expected="BBC Radio 1
Classical Music
Jazz FM
Rock Station
Smooth Jazz"
    
    [ "$result" = "$expected" ]
}

# Test 3: Test jq trimming pattern directly
@test "jq gsub pattern correctly trims whitespace" {
    # Test various whitespace scenarios
    result=$(echo '"  test  "' | jq -r 'gsub("^\\s+|\\s+$";"")')
    [ "$result" = "test" ]
    
    result=$(echo '"  test"' | jq -r 'gsub("^\\s+|\\s+$";"")')
    [ "$result" = "test" ]
    
    result=$(echo '"test  "' | jq -r 'gsub("^\\s+|\\s+$";"")')
    [ "$result" = "test" ]
    
    result=$(echo '"test"' | jq -r 'gsub("^\\s+|\\s+$";"")')
    [ "$result" = "test" ]
}

# Test 4: Test that internal spaces are preserved
@test "internal spaces in station names are preserved" {
    result=$(echo '"  BBC Radio 1  "' | jq -r 'gsub("^\\s+|\\s+$";"")')
    [ "$result" = "BBC Radio 1" ]
    
    result=$(echo '"  My Favorite Station  "' | jq -r 'gsub("^\\s+|\\s+$";"")')
    [ "$result" = "My Favorite Station" ]
}

# Test 5: Test case-insensitive sorting
@test "sorting is case-insensitive" {
    # Create a test file with mixed case names
    cat > "$FAVORITE_PATH/case-test.json" << 'EOF'
[
  {"name": "zebra", "url_resolved": "http://example.com/z", "stationuuid": "z"},
  {"name": "Apple", "url_resolved": "http://example.com/a", "stationuuid": "a"},
  {"name": "Banana", "url_resolved": "http://example.com/b", "stationuuid": "b"},
  {"name": "cherry", "url_resolved": "http://example.com/c", "stationuuid": "c"}
]
EOF
    
    result=$(_station_list "case-test")
    
    expected="Apple
Banana
cherry
zebra"
    
    [ "$result" = "$expected" ]
}

# Test 6: Test empty list
@test "empty station list returns empty string" {
    echo "[]" > "$FAVORITE_PATH/empty-list.json"
    
    result=$(_station_list "empty-list")
    [ -z "$result" ]
}

# Test 7: Test single station
@test "single station is returned correctly" {
    cat > "$FAVORITE_PATH/single.json" << 'EOF'
[
  {
    "name": "  Only Station  ",
    "url_resolved": "http://example.com/only",
    "stationuuid": "only-1"
  }
]
EOF
    
    result=$(_station_list "single")
    [ "$result" = "Only Station" ]
}

# Test 8: Test station names with special characters
@test "special characters in station names are handled correctly" {
    cat > "$FAVORITE_PATH/special.json" << 'EOF'
[
  {
    "name": "  Station & Music  ",
    "url_resolved": "http://example.com/1",
    "stationuuid": "1"
  },
  {
    "name": "Rock 'n' Roll",
    "url_resolved": "http://example.com/2",
    "stationuuid": "2"
  },
  {
    "name": "Jazz @ Night",
    "url_resolved": "http://example.com/3",
    "stationuuid": "3"
  }
]
EOF
    
    result=$(_station_list "special")
    
    expected="Jazz @ Night
Rock 'n' Roll
Station & Music"
    
    [ "$result" = "$expected" ]
}

# Test 9: Test that numbers sort correctly
@test "station names with numbers sort correctly" {
    cat > "$FAVORITE_PATH/numbers.json" << 'EOF'
[
  {
    "name": "Radio 2",
    "url_resolved": "http://example.com/2",
    "stationuuid": "2"
  },
  {
    "name": "Radio 10",
    "url_resolved": "http://example.com/10",
    "stationuuid": "10"
  },
  {
    "name": "Radio 1",
    "url_resolved": "http://example.com/1",
    "stationuuid": "1"
  }
]
EOF
    
    result=$(_station_list "numbers")
    
    # Note: This is alphabetical sorting, not numerical
    # "Radio 1" comes before "Radio 10" which comes before "Radio 2"
    expected="Radio 1
Radio 10
Radio 2"
    
    [ "$result" = "$expected" ]
}

# Test 10: Test very long station names with whitespace
@test "very long station names with whitespace are trimmed" {
    cat > "$FAVORITE_PATH/long.json" << 'EOF'
[
  {
    "name": "  This is a very long radio station name with lots of words in it  ",
    "url_resolved": "http://example.com/long",
    "stationuuid": "long-1"
  }
]
EOF
    
    result=$(_station_list "long")
    [ "$result" = "This is a very long radio station name with lots of words in it" ]
}

# Test 11: Test stations with tabs and other whitespace characters
@test "tabs and other whitespace are trimmed" {
    # Create JSON with escaped tabs (valid JSON format)
    cat > "$FAVORITE_PATH/tabs.json" << 'EOF'
[
  {
    "name": "\tTab Station\t",
    "url_resolved": "http://example.com/tab",
    "stationuuid": "tab-1"
  },
  {
    "name": "  Space Station  ",
    "url_resolved": "http://example.com/space",
    "stationuuid": "space-1"
  }
]
EOF
    
    result=$(_station_list "tabs")
    
    # Both should be trimmed and sorted
    echo "$result" | grep -q "Space Station"
    echo "$result" | grep -q "Tab Station"
}

# Test 12: Test duplicate names (edge case)
@test "duplicate station names are both displayed" {
    cat > "$FAVORITE_PATH/dupes.json" << 'EOF'
[
  {
    "name": "Jazz FM",
    "url_resolved": "http://example.com/jazz1",
    "stationuuid": "jazz-1"
  },
  {
    "name": "Jazz FM",
    "url_resolved": "http://example.com/jazz2",
    "stationuuid": "jazz-2"
  },
  {
    "name": "Rock FM",
    "url_resolved": "http://example.com/rock",
    "stationuuid": "rock-1"
  }
]
EOF
    
    result=$(_station_list "dupes")
    
    # Should have 3 lines
    line_count=$(echo "$result" | wc -l | tr -d ' ')
    [ "$line_count" = "3" ]
    
    # Should contain both Jazz FM entries
    jazz_count=$(echo "$result" | grep -c "Jazz FM")
    [ "$jazz_count" = "2" ]
}

# Test 13: Verify alphabetical order with real-world station names
@test "real-world station names sort correctly" {
    cat > "$FAVORITE_PATH/real.json" << 'EOF'
[
  {
    "name": "  SmoothJazz.com 64k aac+  ",
    "url_resolved": "http://example.com/smooth",
    "stationuuid": "smooth-1"
  },
  {
    "name": "BBC Radio 1",
    "url_resolved": "http://example.com/bbc1",
    "stationuuid": "bbc1"
  },
  {
    "name": "  181.FM - Classic Hits  ",
    "url_resolved": "http://example.com/181",
    "stationuuid": "181"
  },
  {
    "name": "WQXR - New York's Classical Music",
    "url_resolved": "http://example.com/wqxr",
    "stationuuid": "wqxr"
  }
]
EOF
    
    result=$(_station_list "real")
    
    expected="181.FM - Classic Hits
BBC Radio 1
SmoothJazz.com 64k aac+
WQXR - New York's Classical Music"
    
    [ "$result" = "$expected" ]
}

# Test 14: Test that the function handles missing fields gracefully
@test "stations with only name field are handled" {
    cat > "$FAVORITE_PATH/minimal.json" << 'EOF'
[
  {
    "name": "  Minimal Station  "
  },
  {
    "name": "Another Station"
  }
]
EOF
    
    result=$(_station_list "minimal")
    
    expected="Another Station
Minimal Station"
    
    [ "$result" = "$expected" ]
}

# Test 15: Performance test with many stations
@test "handles large lists efficiently" {
    # Create a JSON with 100 stations
    echo "[" > "$FAVORITE_PATH/large.json"
    for i in {1..100}; do
        echo "  {" >> "$FAVORITE_PATH/large.json"
        echo "    \"name\": \"  Station $i  \"," >> "$FAVORITE_PATH/large.json"
        echo "    \"url_resolved\": \"http://example.com/$i\"," >> "$FAVORITE_PATH/large.json"
        echo "    \"stationuuid\": \"uuid-$i\"" >> "$FAVORITE_PATH/large.json"
        if [ $i -lt 100 ]; then
            echo "  }," >> "$FAVORITE_PATH/large.json"
        else
            echo "  }" >> "$FAVORITE_PATH/large.json"
        fi
    done
    echo "]" >> "$FAVORITE_PATH/large.json"
    
    # Just verify it doesn't crash and returns something
    result=$(_station_list "large")
    
    # Should have 100 lines
    line_count=$(echo "$result" | wc -l | tr -d ' ')
    [ "$line_count" = "100" ]
    
    # First line should be "Station 1" (alphabetically first)
    first_line=$(echo "$result" | head -1)
    [ "$first_line" = "Station 1" ]
}
