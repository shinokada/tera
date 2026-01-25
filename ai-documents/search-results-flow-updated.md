## 4. Search Results Screen (UPDATED)

```mermaid
flowchart TD
    Enter([Enter from Search]) --> CheckResults{Results Found?}
    
    CheckResults -->|No| NoResults[Show No Results Message]
    NoResults --> Wait[Wait for Enter]
    Wait --> Back1([Return to Search Menu])
    
    CheckResults -->|Yes| ShowResults[Display Station List - Single Line]
    
    ShowResults --> ResultInput{User Input}
    ResultInput -->|Esc| Back2([Return to Search Menu])
    ResultInput -->|/| Filter[Enter Filter Mode]
    ResultInput -->|Enter| PlayStation[Start MPV Player Immediately]
    
    Filter --> FilterText[Type Filter]
    FilterText --> UpdateResults[Filter Results]
    UpdateResults --> ResultInput
    
    PlayStation --> MPV[Playing Station]
    MPV --> PlayingState{Status}
    PlayingState -->|Stopped| SavePrompt[Show Save Prompt]
    PlayingState -->|q| SavePrompt
    PlayingState -->|s| SaveDuring[Save to Quick Favorites]
    
    SaveDuring --> CheckDupe3{Already Exists?}
    CheckDupe3 -->|Yes| DupeMsg3[Show: Already in Favorites]
    CheckDupe3 -->|No| DoSave3[Add to Quick Favorites]
    DupeMsg3 --> PlayingState
    DoSave3 --> SuccessMsg[Show: Added Successfully]
    SuccessMsg --> PlayingState
    
    SavePrompt --> PromptChoice{Add to Favorites?}
    PromptChoice -->|Yes/1| AddToQuick[Add to My-favorites.json]
    PromptChoice -->|No/2/Esc| Back2
    AddToQuick --> CheckDupe1{Already Exists?}
    CheckDupe1 -->|Yes| DupeMsg1[Show: Already in Favorites]
    CheckDupe1 -->|No| DoAdd[Add Station]
    DupeMsg1 --> Back2
    DoAdd --> Success1[Show Success]
    Success1 --> Back2
```

**State:**
- `results []Station` - Search results from API
- `filteredResults []Station` - After filter applied
- `selectedStation *Station` - Currently selected
- `filterText string` - Current filter

**UI Design:**
- **Search results**: fzf-style display (many results, often 100s-1000s)
- **Single line display**: Format: `NAME • COUNTRY • CODEC BITRATE`
  - Example: `SMOOTH JAZZ • United States • MP3 128kbps`
- Instant filtering with '/' key
- **Direct play on Enter** - No submenu

**Key Logic:**
- Stations displayed in compact single-line format
- **Direct play on Enter** - Immediately starts playback, no intermediate menu
- Check for duplicates by StationUUID
- **Save prompt after playback** - these are NEW discovered stations
- Can also save during playback with 's' key
- Multiple navigation options (Esc to go back)

**Changes from Original:**
- ❌ Removed: Station info submenu (options 1-3)
- ❌ Removed: `i` key for quick info preview
- ✅ Added: Direct play on Enter
- ✅ Changed: Single-line station display
- ✅ Kept: Filter with '/' key
- ✅ Kept: Save to Quick Favorites during/after playback
