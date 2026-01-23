# TERA Screen Flow Charts - Updated

## Summary of Changes
1. **Application Overview & Main Menu**: Removed save prompt after QuickPlay (stations already in My-favorites.json)
2. **Play Screen**: 
   - Context-aware save behavior based on whether station is already in Quick Favorites
   - Use simple arrow navigation for lists, fzf-style for stations
   - Only allow 's' key save during playback if not already in Quick Favorites
3. **UI Display Strategy**: 
   - Lists (few items): Simple arrow navigation
   - Stations & Search Results (many items): fzf-style with filtering

---

## Application Overview

```mermaid
stateDiagram-v2
    [*] --> MainMenu
    MainMenu --> PlayScreen: 1
    MainMenu --> SearchMenu: 2
    MainMenu --> ListMenu: 3
    MainMenu --> DeleteStation: 4
    MainMenu --> Lucky: 5
    MainMenu --> GistMenu: 6
    MainMenu --> QuickPlay: 10-19
    MainMenu --> [*]: 0/q
    
    PlayScreen --> MainMenu: Esc
    SearchMenu --> MainMenu: Esc
    ListMenu --> MainMenu: Esc
    DeleteStation --> MainMenu: Done
    Lucky --> MainMenu: Done
    GistMenu --> MainMenu: Esc
    QuickPlay --> MainMenu: After playback
```

**Note:** QuickPlay stations are from My-favorites.json, so no save prompt after playback.

---

## 1. Main Menu Screen

```mermaid
flowchart TD
    Start([App Start]) --> Init[Initialize Config]
    Init --> LoadFav[Load Quick Favorites]
    LoadFav --> Display[Display Main Menu]
    
    Display --> Input{User Input}
    
    Input -->|1| Play[Navigate to Play Screen]
    Input -->|2| Search[Navigate to Search Menu]
    Input -->|3| List[Navigate to List Menu]
    Input -->|4| Delete[Navigate to Delete Station]
    Input -->|5| Lucky[Play Random Station]
    Input -->|6| Gist[Navigate to Gist Menu]
    Input -->|0/q| Exit([Exit App])
    Input -->|10-19| QuickPlay[Play Quick Favorite]
    
    QuickPlay --> MPV[Start MPV Player]
    MPV --> ShowInfo[Display Station Info]
    ShowInfo --> Playing{Playback Status}
    Playing -->|Stopped| Display
    Playing -->|q pressed| StopMPV[Stop MPV]
    StopMPV --> Display
    
    Play --> Display
    Search --> Display
    List --> Display
    Delete --> Display
    Lucky --> Display
    Gist --> Display
```

**State:**
- `stations []Station` - Quick favorites (My-favorites.json)
- `menuItems []MenuItem` - Dynamic menu with favorites
- `config *Config` - App configuration

**Actions:**
- Load quick favorites on init
- Build dynamic menu items
- Handle numeric shortcuts (10-19)
- Navigate to selected screen
- **No save prompt after QuickPlay** - these stations are already in My-favorites.json

---

## 2. Play Screen

```mermaid
flowchart TD
    Enter([Enter Play Screen]) --> LoadLists[Load All Favorite Lists]
    LoadLists --> ShowLists[Display Lists with Arrow Navigation]
    
    ShowLists --> ListInput{User Input}
    ListInput -->|Esc/0| Back([Return to Main Menu])
    ListInput -->|Select List| LoadStations[Load Stations from List]
    
    LoadStations --> ShowStations[Display Stations with fzf-style]
    ShowStations --> StationInput{User Input}
    
    StationInput -->|Esc/0| ShowLists
    StationInput -->|/| Filter[Enter Filter Mode]
    StationInput -->|Select| GetStation[Get Station Data]
    
    Filter --> FilterInput[Type Filter Text]
    FilterInput --> UpdateList[Filter Station List]
    UpdateList --> StationInput
    
    GetStation --> CheckInQuick{Already in Quick Favorites?}
    CheckInQuick -->|Yes| ShowInfo1[Display Station Info]
    CheckInQuick -->|No| ShowInfo2[Display Station Info + Save Option]
    
    ShowInfo1 --> StartPlayer1[Start MPV Player]
    ShowInfo2 --> StartPlayer2[Start MPV Player]
    
    StartPlayer1 --> Playing1{Playback}
    StartPlayer2 --> Playing2{Playback}
    
    Playing1 -->|q| Stop1[Stop Playback]
    Playing1 -->|Error| Error1[Show Error Message]
    
    Playing2 -->|q| Stop2[Stop Playback]
    Playing2 -->|s| SaveToQuick[Save to Quick Favorites]
    Playing2 -->|Error| Error2[Show Error Message]
    
    Stop1 --> ShowLists
    Error1 --> ShowLists
    Stop2 --> ShowLists
    Error2 --> ShowLists
    
    SaveToQuick --> CheckDupe{Duplicate?}
    CheckDupe -->|Yes| AlreadyMsg[Show: Already in Quick Favorites]
    CheckDupe -->|No| DoSave[Add to Quick Favorites]
    AlreadyMsg --> Continue[Continue Playing]
    DoSave --> SuccessMsg[Show: Added to Quick Favorites]
    SuccessMsg --> Continue
    Continue --> Playing2
```

**State:**
- `lists []string` - Available favorite lists
- `selectedList string` - Currently selected list
- `stations []Station` - Stations in selected list
- `filterText string` - Current filter
- `player *MPVPlayer` - Player instance

**UI Design:**
- **Lists**: Simple arrow key navigation (few items, typically 3-10 lists)
- **Stations**: fzf-style with filter capability (moderate items, 10-100 stations)
  - Provides quick filtering even for smaller lists
  - Consistent user experience across the app
  - Stations sorted alphabetically (case-insensitive)

**Key Logic:**
- Check if station is already in Quick Favorites (My-favorites.json) by StationUUID
- If already in Quick Favorites: Don't show save option (no 's' key, no prompt after)
- If from another list: Allow saving to Quick Favorites during playback (press 's')
- Check for duplicates by StationUUID before adding
- **No save prompt after playback** - only during playback with 's' key
- If user presses 's' but station is already in Quick Favorites, show friendly message

**Rationale:**
- Stations from My-favorites.json → Already saved, no need to save again
- Stations from other lists → User might want to promote to Quick Favorites for main menu access
- Simple, clear UX without redundant prompts

---

## 3. Search Menu Screen

```mermaid
flowchart TD
    Enter([Enter Search Menu]) --> ShowMenu[Display Search Options]
    
    ShowMenu --> MenuInput{User Input}
    MenuInput -->|0/Esc| Back([Return to Main Menu])
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
    
    NavCheck1 -->|0| ShowMenu
    NavCheck1 -->|00| Back
    NavCheck1 -->|Query| DoSearch1[Call API]
    
    NavCheck2 -->|0| ShowMenu
    NavCheck2 -->|00| Back
    NavCheck2 -->|Query| DoSearch2[Call API]
    
    NavCheck3 -->|0| ShowMenu
    NavCheck3 -->|00| Back
    NavCheck3 -->|Query| DoSearch3[Call API]
    
    NavCheck4 -->|0| ShowMenu
    NavCheck4 -->|00| Back
    NavCheck4 -->|Query| DoSearch4[Call API]
    
    NavCheck5 -->|0| ShowMenu
    NavCheck5 -->|00| Back
    NavCheck5 -->|Query| DoSearch5[Call API]
    
    NavCheck6 -->|0| ShowMenu
    NavCheck6 -->|00| Back
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

**State:**
- `searchType SearchType` - Tag, Name, Language, Country, State, Advanced
- `query string` - User's search query
- `loading bool` - API call in progress

**Actions:**
- POST to Radio Browser API
- Handle navigation shortcuts (0, 00)
- Display spinner during search
- Navigate to results on success

---

## 4. Search Results Screen

```mermaid
flowchart TD
    Enter([Enter from Search]) --> CheckResults{Results Found?}
    
    CheckResults -->|No| NoResults[Show No Results Message]
    NoResults --> Wait[Wait for Enter]
    Wait --> Back1([Return to Search Menu])
    
    CheckResults -->|Yes| ShowResults[Display Station List with fzf-style]
    
    ShowResults --> ResultInput{User Input}
    ResultInput -->|Esc| Back2([Return to Search Menu])
    ResultInput -->|/| Filter[Enter Filter Mode]
    ResultInput -->|i| QuickInfo[Show Info Overlay]
    ResultInput -->|Select| ShowStationInfo[Display Full Station Info]
    
    Filter --> FilterText[Type Filter]
    FilterText --> UpdateResults[Filter Results]
    UpdateResults --> ResultInput
    
    QuickInfo --> ResultInput
    
    ShowStationInfo --> Submenu[Show Station Submenu]
    Submenu --> SubInput{User Input}
    
    SubInput -->|0| Back2
    SubInput -->|1: Play| PlayStation[Start MPV Player]
    SubInput -->|2: Save| ShowSaveLists[Show Lists to Save To]
    SubInput -->|3: Back| ShowResults
    
    PlayStation --> MPV[Playing Station]
    MPV --> PlayingState{Status}
    PlayingState -->|Stopped| SavePrompt[Show Save Prompt]
    PlayingState -->|q| SavePrompt
    SavePrompt --> PromptChoice{Add to Favorites?}
    PromptChoice -->|Yes| AddToQuick[Add to My-favorites.json]
    PromptChoice -->|No| Back2
    AddToQuick --> CheckDupe1{Already Exists?}
    CheckDupe1 -->|Yes| DupeMsg1[Show: Already in Favorites]
    CheckDupe1 -->|No| DoAdd[Add Station]
    DupeMsg1 --> Back2
    DoAdd --> Success1[Show Success]
    Success1 --> Back2
    
    ShowSaveLists --> ListChoice{Select List}
    ListChoice -->|0| Submenu
    ListChoice -->|00| Back2
    ListChoice -->|Select| CheckDupe2{Duplicate?}
    CheckDupe2 -->|Yes| DupeMsg2[Show: Already in List]
    CheckDupe2 -->|No| SaveIt[Save to List]
    DupeMsg2 --> Back2
    SaveIt --> Success2[Show Success]
    Success2 --> Back2
```

**State:**
- `results []Station` - Search results from API
- `filteredResults []Station` - After filter applied
- `selectedStation *Station` - Currently selected
- `filterText string` - Current filter

**UI Design:**
- **Search results**: fzf-style display (many results, often 100s-1000s)
- Instant filtering with '/' key
- Quick info preview with 'i' key

**Key Logic:**
- Check for duplicates by StationUUID
- **Save prompt after playback** - these are NEW discovered stations
- Multiple navigation options (0, 00, Esc)
- Two save paths:
  1. Play then auto-prompt for Quick Favorites
  2. Save to any list without playing

---

## 5. List Management Menu Screen
[Same as original - no changes]

---

## 6. Delete Station Screen

```mermaid
flowchart TD
    Enter([Enter Delete Station]) --> LoadLists[Load All Lists]
    LoadLists --> ShowLists[Display Lists with Arrow Navigation]
    
    ShowLists --> ListInput{Select List}
    ListInput -->|Esc/0| Back([Return to Main Menu])
    ListInput -->|Select| LoadStations[Load Stations from List]
    
    LoadStations --> ShowStations[Display Stations with fzf-style]
    ShowStations --> StationInput{Select Station}
    
    StationInput -->|Esc/0| ShowLists
    StationInput -->|/| Filter[Enter Filter Mode]
    StationInput -->|Select| Confirm[Show Confirmation]
    
    Filter --> FilterText[Type Filter]
    FilterText --> UpdateList[Filter Station List]
    UpdateList --> StationInput
    
    Confirm --> ConfirmInput{Confirm Delete?}
    ConfirmInput -->|No/Esc| ShowStations
    ConfirmInput -->|Yes| DoDelete[Remove from JSON]
    DoDelete --> SaveFile[Save Updated File]
    SaveFile --> Success[Show Success]
    Success --> CheckEmpty{List Empty?}
    CheckEmpty -->|Yes| ShowLists
    CheckEmpty -->|No| ShowStations
```

**State:**
- `lists []string` - Available lists
- `selectedList string` - Current list
- `stations []Station` - Stations in list
- `selectedStation *Station` - Station to delete

**UI Design:**
- **Lists**: Simple arrow navigation (few items)
- **Stations**: fzf-style with filtering (many items)

**Actions:**
- Find station by StationUUID
- Remove from array
- Save file atomically
- Handle empty list state

---

## 7. Lucky Screen (I Feel Lucky)

```mermaid
flowchart TD
    Enter([Enter Lucky]) --> ShowMsg[Display: Finding Random Station...]
    ShowMsg --> CallAPI[Call Radio Browser API]
    CallAPI --> GetRandom[Get Random Popular Station]
    
    GetRandom --> APIResponse{API Success?}
    APIResponse -->|No| Error[Show Error Message]
    APIResponse -->|Yes| ShowInfo[Display Station Info]
    
    Error --> Back1([Return to Main Menu])
    
    ShowInfo --> StartPlayer[Start MPV Player]
    StartPlayer --> Playing{Playback}
    
    Playing -->|q| Stop[Stop Playback]
    Playing -->|s| SaveToQuick[Save to Quick Favorites]
    Playing -->|Error| PlayError[Show Error]
    
    Stop --> Prompt[Show Save Prompt]
    PlayError --> Prompt
    SaveToQuick --> CheckDupe1{Already Exists?}
    CheckDupe1 -->|Yes| AlreadyMsg[Show: Already Saved]
    CheckDupe1 -->|No| DoSave[Add to Quick Favorites]
    AlreadyMsg --> Continue[Continue Playing]
    DoSave --> SuccessMsg[Show: Added Successfully]
    SuccessMsg --> Continue
    Continue --> Playing
    
    Prompt --> PromptChoice{Save to Favorites?}
    PromptChoice -->|Yes| AddToQuick[Add to My-favorites.json]
    PromptChoice -->|No| Back2([Return to Main Menu])
    AddToQuick --> CheckDupe2{Already Exists?}
    CheckDupe2 -->|Yes| AlreadyMsg2[Show: Already Saved]
    CheckDupe2 -->|No| DoAdd[Add Station]
    AlreadyMsg2 --> Back2
    DoAdd --> Success[Show Success]
    Success --> Back2
```

**State:**
- `station *Station` - Random station selected
- `player *MPVPlayer` - Player instance

**Logic:**
- Query API for high-vote stations
- Select random from results
- **Save prompt after playback** (NEW discovery)
- Can also save during playback with 's' key
- Duplicate checking for both save methods

---

## 8-14. Remaining Screens
[Gist Menu, Create Gist, My Gists, Token Management, Update Gist, Delete Gist, Recover from Gist - no changes from original]

---

## UI Display Guidelines Summary

**When to use Simple Arrow Navigation:**
- Favorite lists selection (typically 3-10 items)
- Menu options (fixed, small set)
- Gist lists (typically 1-10 items)
- Any list with < 15 items where user knows what they're looking for

**When to use fzf-style with Filtering:**
- Radio station lists from search (100s-1000s of results)
- Stations within a favorite list (10-100 stations)
- Any list where quick filtering is beneficial
- Any list with > 15 items

**Benefits of this approach:**
- Simple navigation where it makes sense (don't overcomplicate)
- Powerful filtering where it's needed
- Consistent experience for similar types of content
- Better performance (don't run fzf for 3 items)
