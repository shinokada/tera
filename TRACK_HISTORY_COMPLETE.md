# Track History Feature - Implementation Complete! 🎉

## ✅ What Was Implemented

I've successfully added the track history feature to display the last 5 track names while playing a radio station!

### Changes Made

#### 1. **internal/player/mpv.go**
- Added track history fields to `MPVPlayer` struct:
  - `trackHistory []string` - Last 5 track names
  - `currentTrack string` - Current playing track
  - `trackMu sync.Mutex` - Thread-safe access
  
- Added new methods:
  - `GetCurrentTrack()` - Retrieves current track from MPV metadata
  - `GetTrackHistory()` - Returns copy of last 5 tracks
  - `addToTrackHistory()` - Adds new track to history (smart deduplication)
  - `monitorMetadata()` - Background goroutine that polls every 5 seconds
  
- Modified existing methods:
  - `Play()` - Now starts metadata monitoring
  - `stopInternal()` - Clears track history on stop

#### 2. **internal/ui/play.go**
- Added `trackHistory []string` field to `PlayModel`
- Added `trackHistoryMsg` message type
- Added `pollTrackHistory()` command function
- Updated `Update()` method to handle track updates
- Updated `viewPlaying()` to display track history with styling
- Clear track history when stopping playback

### How It Works

1. **Metadata Polling**: MPV monitors `media-title` property every 5 seconds
2. **Smart Tracking**: Only adds new tracks (deduplicates current track)
3. **History Management**: Maintains last 5 unique tracks (newest first)
4. **UI Display**: Shows tracks with 🎵 indicator for current track
5. **Auto-Clear**: Clears history when playback stops

### What You'll See

When playing a station that sends track metadata:

```
🎵 Now Playing
────────────────────────────────────────
Station: Jazz FM
Country: United States
Codec: MP3 128kbps

▶ Playing...

Recently Played:
🎵 Miles Davis - So What
   John Coltrane - Blue Train
   Bill Evans - Waltz for Debby
   Thelonious Monk - Round Midnight
   Chet Baker - My Funny Valentine
```

### Features

✅ Non-blocking 5-second polling interval  
✅ Thread-safe with proper mutex protection  
✅ Automatic deduplication of tracks  
✅ Shows max 5 tracks (configurable)  
✅ Current track highlighted with 🎵  
✅ Auto-clears on stop  
✅ Color-coded (current = blue, older = gray)  
✅ Only polls when actively playing  

### Build and Test

```bash
cd /Users/shinichiokada/Terminal-Tools/tera
make clean-all && make lint && make build
./tera
```

Then:
1. Navigate to "Play from Favorites"
2. Select a favorite list
3. Play a radio station
4. Wait a few minutes for track changes
5. Watch the "Recently Played" section update!

### Important Notes

- **Not all stations send metadata** - Some stations don't provide `icy-title` metadata
- **Update frequency varies** - Depends on when station updates metadata (typically 3-5 minutes per song)
- **Track format varies** - Some stations send "Artist - Song", others just "Song Title"
- **Minimum length filter** - Tracks shorter than 3 characters are filtered out (likely station IDs)

### Example Stations That Send Metadata

Most internet radio stations send track info, especially:
- Music-focused stations (Jazz, Classical, Rock, Pop)
- Streaming services (SomaFM, Radio Paradise, etc.)
- Public radio music shows

Stations that typically DON'T send metadata:
- Talk radio stations
- News stations
- Some smaller local stations

Enjoy your new track history feature! 🎵
