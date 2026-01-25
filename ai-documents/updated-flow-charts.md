# Updated Flow Charts - Keyboard Shortcuts Standardization

## Key Changes Throughout All Flow Charts

### Global Changes
- **Removed:** `0` (back), `00` (main menu)
- **Standardized:** `Esc` (back), `q` (quit)

---

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
    ResultInput -->|q| Quit([Quit Application])
    ResultInput -->|Enter| PlayStation[Start MPV Player Immediately]
    
    PlayStation --> MPV[Playing Station]
    MPV --> PlayingState{Status}
    PlayingState -->|q| SavePrompt[Show Save Prompt]
    PlayingState -->|Esc| StopNoSave[Stop & Back to Results]
    PlayingState -->|s| SaveDuring[Save to Quick Favorites]
    
    StopNoSave --> Back2
    
    SaveDuring --> CheckDupe3{Already Exists?}
    CheckDupe3 -->|Yes| DupeMsg3[Show: Already in Favorites]
    CheckDupe3 -->|No| DoSave3[Add to Quick Favorites]
    DupeMsg3 --> PlayingState
    DoSave3 --> SuccessMsg[Show: Added Successfully]
    SuccessMsg --> PlayingState
    
    SavePrompt --> PromptChoice{Add to Favorites?}
    PromptChoice -->|y/1| AddToQuick[Add to My-favorites.json]
    PromptChoice -->|n/2/Esc| Back2
    PromptChoice -->|q| Quit
    
    AddToQuick --> CheckDupe1{Already Exists?}
    CheckDupe1 -->|Yes| DupeMsg1[Show: Already in Favorites]
    CheckDupe1 -->|No| DoAdd[Add Station]
    DupeMsg1 --> Back2
    DoAdd --> Success1[Show Success]
    Success1 --> Back2
```

**Key Updates:**
- Added `q` → Quit from results list
- Added `q` → Quit from save prompt
- Separated `q` and `Esc` behavior during playback:
  - `q` - Shows save prompt
  - `Esc` - Goes back without prompt
- Removed all `0` and `00` references

---

## 2. Play Screen (UPDATED)

```mermaid
flowchart TD
    Enter([Enter Play Screen]) --> LoadLists[Load All Favorite Lists]
    LoadLists --> ShowLists[Display Lists with Arrow Navigation]
    
    ShowLists --> ListInput{User Input}
    ListInput -->|Esc| Back([Return to Main Menu])
    ListInput -->|q| Quit([Quit Application])
    ListInput -->|Select List| LoadStations[Load Stations from List]
    
    LoadStations --> ShowStations[Display Stations with fzf-style]
    ShowStations --> StationInput{User Input}
    
    StationInput -->|Esc| ShowLists
    StationInput -->|q| Quit
    StationInput -->|Select| GetStation[Get Station Data]
    
    GetStation --> CheckInQuick{Already in Quick Favorites?}
    CheckInQuick -->|Yes| ShowInfo1[Display Station Info]
    CheckInQuick -->|No| ShowInfo2[Display Station Info + Save Option]
    
    ShowInfo1 --> StartPlayer1[Start MPV Player]
    ShowInfo2 --> StartPlayer2[Start MPV Player]
    
    StartPlayer1 --> Playing1{Playback}
    StartPlayer2 --> Playing2{Playback}
    
    Playing1 -->|q or Esc| Stop1[Stop Playback]
    Playing2 -->|q or Esc| Stop2[Stop Playback]
    Playing2 -->|s| SaveToQuick[Save to Quick Favorites]
    
    Stop1 --> ShowLists
    Stop2 --> ShowLists
    
    SaveToQuick --> CheckDupe{Duplicate?}
    CheckDupe -->|Yes| AlreadyMsg[Show: Already in Quick Favorites]
    CheckDupe -->|No| DoSave[Add to Quick Favorites]
    AlreadyMsg --> Continue[Continue Playing]
    DoSave --> SuccessMsg[Show: Added to Quick Favorites]
    SuccessMsg --> Continue
    Continue --> Playing2
```

**Key Updates:**
- Changed `Esc/0` → `Esc` for back
- Added `q` for quit at each level
- Removed `0` shortcuts
- Combined `q/esc/0` → `q or Esc` in playing state

---

## 3. Search Menu Screen (UPDATED)

```mermaid
flowchart TD
    Enter([Enter Search Menu]) --> ShowMenu[Display Search Options]
    
    ShowMenu --> MenuInput{User Input}
    MenuInput -->|Esc| Back([Return to Main Menu])
    MenuInput -->|q| Quit([Quit Application])
    MenuInput -->|1| Tag[Search by Tag]
    MenuInput -->|2| Name[Search by Name]
    MenuInput -->|3| Lang[Search by Language]
    MenuInput -->|4| Country[Search by Country Code]
    MenuInput -->|5| State[Search by State]
    MenuInput -->|6| Advanced[Advanced Search]
    
    Tag --> TagInput[Prompt: Enter Tag]
    Name --> NameInput[Prompt: Enter Name]
    Lang --> LangInput[Prompt: Enter Language]
    Country --> CountryInput[Prompt: Enter Country Code]
    State --> StateInput[Prompt: Enter State]
    Advanced --> AdvInput[Prompt: Enter Multiple Criteria]
    
    TagInput --> NavCheck1{Input}
    NameInput --> NavCheck2{Input}
    LangInput --> NavCheck3{Input}
    CountryInput --> NavCheck4{Input}
    StateInput --> NavCheck5{Input}
    AdvInput --> NavCheck6{Input}
    
    NavCheck1 -->|Esc| ShowMenu
    NavCheck1 -->|q| Quit
    NavCheck1 -->|Query| DoSearch1[Call API]
    
    NavCheck2 -->|Esc| ShowMenu
    NavCheck2 -->|q| Quit
    NavCheck2 -->|Query| DoSearch2[Call API]
    
    NavCheck3 -->|Esc| ShowMenu
    NavCheck3 -->|q| Quit
    NavCheck3 -->|Query| DoSearch3[Call API]
    
    NavCheck4 -->|Esc| ShowMenu
    NavCheck4 -->|q| Quit
    NavCheck4 -->|Query| DoSearch4[Call API]
    
    NavCheck5 -->|Esc| ShowMenu
    NavCheck5 -->|q| Quit
    NavCheck5 -->|Query| DoSearch5[Call API]
    
    NavCheck6 -->|Esc| ShowMenu
    NavCheck6 -->|q| Quit
    NavCheck6 -->|Query| DoSearch6[Call API]
    
    DoSearch1 --> Loading1[Show Searching...]
    DoSearch2 --> Loading2[Show Searching...]
    DoSearch3 --> Loading3[Show Searching...]
    DoSearch4 --> Loading4[Show Searching...]
    DoSearch5 --> Loading5[Show Searching...]
    DoSearch6 --> Loading6[Show Searching...]
    
    Loading1 --> Results[Navigate to Search Results]
    Loading2 --> Results
    Loading3 --> Results
    Loading4 --> Results
    Loading5 --> Results
    Loading6 --> Results
```

**Key Updates:**
- Changed `0` → `Esc` for back to menu
- Changed `00` → `q` for quit
- Added `q` → Quit at menu level
- Removed all `0` and `00` references

---

## Summary of Flow Chart Updates

### Every Flow Chart Needs:
1. Replace `0` with `Esc` for "back one level"
2. Replace `00` with `q` for "quit application"  
3. Add `q` option at every interactive state
4. Remove `0/Esc` combined options → just `Esc`
5. Update help text in all states

### Standard Navigation Pattern:
```text
Any Screen
├── Esc → Back one level
├── q → Quit application
└── Ctrl+C → Force quit
```

This provides consistent, industry-standard navigation throughout the entire application.
