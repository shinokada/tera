# TERA Screen Flow Charts


## Application Overview

```mermaid
stateDiagram-v2
    [*] --> MainMenu
    MainMenu --> PlayScreen: 1
    MainMenu --> SearchMenu: 2
    MainMenu --> ListMenu: 3
    MainMenu --> Lucky: 4
    MainMenu --> GistMenu: 5
    MainMenu --> QuickPlay: 10-19
    MainMenu --> [*]: 0/Ctrl+C
    
    PlayScreen --> MainMenu: Esc
    SearchMenu --> MainMenu: Esc
    ListMenu --> MainMenu: Esc
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
    Input -->|4| Lucky[Play Random Station]
    Input -->|5| Gist[Navigate to Gist Menu]
    Input -->|0/Ctrl+C| Exit([Exit App])
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
    StationInput -->|d| ConfirmDelete{Confirm Delete?}
    StationInput -->|i| InfoPreview[Show Info Overlay]
    StationInput -->|Select| GetStation[Get Station Data]
    
    Filter --> FilterInput[Type Filter Text]
    FilterInput --> UpdateList[Filter Station List]
    UpdateList --> StationInput
    
    ConfirmDelete -->|No/Esc| ShowStations
    ConfirmDelete -->|Yes| DeleteStation[Remove from List]
    DeleteStation --> SaveJSON[Save Updated JSON]
    SaveJSON --> DeleteSuccess[Show Success Message]
    DeleteSuccess --> CheckEmpty{List Empty?}
    CheckEmpty -->|Yes| ShowLists
    CheckEmpty -->|No| ShowStations
    
    InfoPreview -->|Esc/i| ShowStations
    
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
    ResultInput -->|0: Back| ShowResults
    ResultInput -->|1: Play| PlayStation[Start MPV Player]
    
    PlayStation --> MPV[Playing Station with Staion status or play list from station]
    MPV -->|q| ShowSaveLists
    
    ShowSaveLists --> ListChoice{Select List}
    ListChoice -->|Esc| Back2
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

**Key Logic:**
- Check for duplicates by StationUUID
- **Save prompt after playback** - these are NEW discovered stations
- Multiple navigation options (0, 00, Esc)
- Two save paths:
  1. Play then auto-prompt for Quick Favorites
  2. Save to any list without playing

---

## 5. List Management Menu Screen

```mermaid
flowchart TD
    Enter([Enter List Menu]) --> ShowMenu[Display List Management Options]
    
    ShowMenu --> MenuInput{User Input}
    MenuInput -->|0/Esc| Back([Return to Main Menu])
    MenuInput -->|1| Create[Create New List]
    MenuInput -->|2| Delete[Delete List]
    MenuInput -->|3| Edit[Edit List Name]
    MenuInput -->|4| ShowAll[Show All Lists]
    
    Create --> ShowLists1[Display Current Lists]
    ShowLists1 --> NameInput1[Prompt: Enter New Name]
    NameInput1 --> NavCheck1{Input}
    NavCheck1 -->|0| ShowMenu
    NavCheck1 -->|00| Back
    NavCheck1 -->|Empty| Error1[Show Error: Name Required]
    NavCheck1 -->|Name| CheckExists1{List Exists?}
    Error1 --> Create
    CheckExists1 -->|Yes| Error2[Show Error: Already Exists]
    CheckExists1 -->|No| DoCreate[Create List File]
    Error2 --> Create
    DoCreate --> InitFile[Initialize with Empty Array]
    InitFile --> Success1[Show Success Message]
    Success1 --> ShowMenu
    
    Delete --> ShowLists2[Display Current Lists]
    ShowLists2 --> NameInput2[Prompt: Enter Name to Delete]
    NameInput2 --> NavCheck2{Input}
    NavCheck2 -->|0| ShowMenu
    NavCheck2 -->|00| Back
    NavCheck2 -->|Empty| Error3[Show Error: Name Required]
    NavCheck2 -->|Name| CheckProtected{Protected List?}
    Error3 --> Delete
    CheckProtected -->|Yes: My-favorites| Error4[Cannot Delete My-favorites]
    CheckProtected -->|No| CheckExists2{List Exists?}
    Error4 --> Delete
    CheckExists2 -->|No| Error5[List Doesn't Exist]
    CheckExists2 -->|Yes| DoDelete[Delete File]
    Error5 --> Delete
    DoDelete --> Success2[Show Success]
    Success2 --> ShowMenu
    
    Edit --> ShowLists3[Display Current Lists]
    ShowLists3 --> NameInput3[Prompt: Enter Name to Edit]
    NameInput3 --> NavCheck3{Input}
    NavCheck3 -->|0| ShowMenu
    NavCheck3 -->|00| Back
    NavCheck3 -->|Empty| Error6[Show Error: Name Required]
    NavCheck3 -->|Name| CheckExists3{List Exists?}
    Error6 --> Edit
    CheckExists3 -->|No| Error7[List Doesn't Exist]
    CheckExists3 -->|Yes| CheckProtected2{Protected?}
    Error7 --> Edit
    CheckProtected2 -->|Yes| Error8[Cannot Rename My-favorites]
    CheckProtected2 -->|No| NewNameInput[Prompt: Enter New Name]
    Error8 --> Edit
    NewNameInput --> NavCheck4{Input}
    NavCheck4 -->|0| ShowMenu
    NavCheck4 -->|00| Back
    NavCheck4 -->|Empty| Error9[Name Required]
    NavCheck4 -->|Name| CheckExists4{New Name Exists?}
    Error9 --> NewNameInput
    CheckExists4 -->|Yes| Error10[Name Already Taken]
    CheckExists4 -->|No| DoRename[Rename File]
    Error10 --> NewNameInput
    DoRename --> Success3[Show Success]
    Success3 --> ShowMenu
    
    ShowAll --> ListAll[Display All List Names]
    ListAll --> WaitEnter[Wait for Enter]
    WaitEnter --> ShowMenu
```

**State:**
- `lists []string` - Available lists
- `operation Operation` - Create, Delete, Edit, ShowAll
- `inputValue string` - User input

**Validation Rules:**
- List names cannot be empty
- Names must be unique
- "My-favorites" is protected (cannot delete/rename)
- Replace spaces with hyphens in names

---

## 6. Lucky Screen (I Feel Lucky)

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

## 7. Gist Menu Screen

```mermaid
flowchart TD
    Enter([Enter Gist Menu]) --> CheckToken{Token Configured?}
    CheckToken -->|No| ShowWarning[Display Warning: No Token]
    CheckToken -->|Yes| ShowStatus[Display Token Status]
    
    ShowWarning --> ShowMenu[Display Gist Menu]
    ShowStatus --> ShowMenu
    
    ShowMenu --> MenuInput{User Input}
    MenuInput -->|0/Esc| Back([Return to Main Menu])
    MenuInput -->|1| Create[Create Gist]
    MenuInput -->|2| MyGists[My Gists]
    MenuInput -->|3| Recover[Recover from Gist]
    MenuInput -->|4| Update[Update Gist]
    MenuInput -->|5| Delete[Delete Gist]
    MenuInput -->|6| Token[Token Management]
    
    Create --> CreateFlow[Navigate to Create Gist]
    MyGists --> MyGistsFlow[Navigate to My Gists]
    Recover --> RecoverFlow[Navigate to Recover]
    Update --> UpdateFlow[Navigate to Update Gist]
    Delete --> DeleteFlow[Navigate to Delete Gist]
    Token --> TokenFlow[Navigate to Token Management]
    
    CreateFlow --> ShowMenu
    MyGistsFlow --> ShowMenu
    RecoverFlow --> ShowMenu
    UpdateFlow --> ShowMenu
    DeleteFlow --> ShowMenu
    TokenFlow --> ShowMenu
```

**State:**
- `hasToken bool` - Token configured status
- `gistCount int` - Number of saved gists
- `currentToken string` - Masked token display

---

## 8. Create Gist Screen

```mermaid
flowchart TD
    Enter([Enter Create Gist]) --> CheckToken{Token Available?}
    
    CheckToken -->|No| TokenError[Show: Token Not Found]
    TokenError --> Instructions[Display Setup Instructions]
    Instructions --> Back1([Return to Gist Menu])
    
    CheckToken -->|Yes| LoadLists[Load All Favorite Lists]
    LoadLists --> CheckLists{Lists Available?}
    
    CheckLists -->|No| NoListsError[Show: No Lists Found]
    NoListsError --> HelpMsg[Suggest: Create Lists First]
    HelpMsg --> Back2([Return to Gist Menu])
    
    CheckLists -->|Yes| PrepareFiles[Prepare Files for Upload]
    PrepareFiles --> BuildJSON[Build JSON Payload]
    BuildJSON --> ShowProgress[Display: Creating Gist...]
    ShowProgress --> CallGitHub[POST to GitHub API]
    
    CallGitHub --> APIResponse{Success?}
    APIResponse -->|No| Error[Show Error Details]
    Error --> Troubleshoot[Display Common Issues]
    Troubleshoot --> Back3([Return to Gist Menu])
    
    APIResponse -->|Yes| SaveMetadata[Save Gist Metadata Locally]
    SaveMetadata --> ShowSuccess[Display Success with URL]
    ShowSuccess --> OpenBrowser[Open in Browser]
    OpenBrowser --> Back4([Return to Gist Menu])
```

**State:**
- `token string` - GitHub token
- `files []FavoritesList` - Files to upload
- `gistID string` - Created gist ID
- `gistURL string` - Created gist URL

**Actions:**
- Read all JSON files from favorite path
- Build GitHub gist payload
- Save gist metadata for tracking
- Handle errors with helpful messages

---

## 9. My Gists Screen

```mermaid
flowchart TD
    Enter([Enter My Gists]) --> LoadMetadata[Load Local Gist Metadata]
    LoadMetadata --> CheckCount{Gists Available?}
    
    CheckCount -->|No| NoGists[Show: No Gists Created]
    NoGists --> Suggest[Suggest: Create First Gist]
    Suggest --> Back1([Return to Gist Menu])
    
    CheckCount -->|Yes| ShowList[Display Gist List]
    ShowList --> ListFormat[Format: Description | Created Date]
    ListFormat --> UserInput{User Input}
    
    UserInput -->|0/Esc| Back2([Return to Gist Menu])
    UserInput -->|Select Number| GetGist[Get Selected Gist]
    
    GetGist --> OpenURL[Open Gist URL in Browser]
    OpenURL --> Wait[Wait for Enter]
    Wait --> ShowList
```

**State:**
- `gists []GistMetadata` - Local gist records
- `selectedGist *GistMetadata` - User selection

**Display:**
- Index | Description | Created Date
- Up to 10 gists per page

---

## 10. Token Management Screen

```mermaid
flowchart TD
    Enter([Enter Token Management]) --> CheckStatus{Token Exists?}
    
    CheckStatus -->|Yes| ShowCurrent[Display Current Token Masked]
    CheckStatus -->|No| ShowNone[Display: No Token Configured]
    
    ShowCurrent --> ShowMenu[Display Token Menu]
    ShowNone --> ShowMenu
    
    ShowMenu --> MenuInput{User Input}
    MenuInput -->|0/Esc| Back([Return to Gist Menu])
    MenuInput -->|1| Setup[Setup/Change Token]
    MenuInput -->|2| View[View Current Token]
    MenuInput -->|3| Validate[Validate Token]
    MenuInput -->|4| DeleteToken[Delete Token]
    
    Setup --> Instructions[Show GitHub Token Instructions]
    Instructions --> TokenInput[Prompt: Paste Token Hidden]
    TokenInput --> BasicCheck{Valid Format?}
    BasicCheck -->|No: Too Short| Error1[Show Format Error]
    BasicCheck -->|Yes| APIValidate[Validate with GitHub API]
    Error1 --> Setup
    APIValidate --> ValidResponse{Valid?}
    ValidResponse -->|No| Error2[Show Invalid Token Error]
    ValidResponse -->|Yes| GetUsername[Get GitHub Username]
    Error2 --> Setup
    GetUsername --> SaveToken[Save to Keyring/File]
    SaveToken --> Success1[Show Success + Username]
    Success1 --> ShowMenu
    
    View --> LoadToken[Load Current Token]
    LoadToken --> MaskToken[Mask Token Display]
    MaskToken --> CheckValid[Validate with API]
    CheckValid --> ValidStatus{Valid?}
    ValidStatus -->|Yes| ShowValid[Display: Token Valid]
    ValidStatus -->|No| ShowInvalid[Display: Token Invalid]
    ShowValid --> Back
    ShowInvalid --> Back
    
    Validate --> TestAPI[Call GitHub API]
    TestAPI --> TestResponse{Success?}
    TestResponse -->|Yes| ValidMsg[Display: Valid + Username]
    TestResponse -->|No| InvalidMsg[Display: Invalid + Reasons]
    ValidMsg --> Back
    InvalidMsg --> Back
    
    DeleteToken --> ConfirmDelete[Prompt: Confirm Deletion]
    ConfirmDelete --> ConfirmInput{Confirm?}
    ConfirmInput -->|No| ShowMenu
    ConfirmInput -->|Yes: type yes| DoDelete[Delete from Storage]
    DoDelete --> ClearEnv[Clear Environment Variable]
    ClearEnv --> Success2[Show Success]
    Success2 --> ShowMenu
```

**State:**
- `token string` - Current token
- `maskedToken string` - Display version
- `username string` - GitHub username
- `valid bool` - Validation status

**Security:**
- Hidden input when typing token
- Mask all displays (show first 11 + last 4)
- Store in keyring (fallback to encrypted file)
- Clear from memory after operations

---

## 11. Update Gist Screen

```mermaid
flowchart TD
    Enter([Enter Update Gist]) --> CheckToken{Token Available?}
    
    CheckToken -->|No| TokenError[Show: Token Required]
    TokenError --> Back1([Return to Gist Menu])
    
    CheckToken -->|Yes| LoadMetadata[Load Gist Metadata]
    LoadMetadata --> CheckGists{Gists Available?}
    
    CheckGists -->|No| NoGists[Show: No Gists to Update]
    NoGists --> Back2([Return to Gist Menu])
    
    CheckGists -->|Yes| ShowList[Display Gist List]
    ShowList --> UserInput{User Input}
    
    UserInput -->|0/Esc| Back3([Return to Gist Menu])
    UserInput -->|Select| GetGist[Get Selected Gist]
    
    GetGist --> ShowCurrent[Display Current Description]
    ShowCurrent --> PromptNew[Prompt: Enter New Description]
    
    PromptNew --> DescInput{User Input}
    DescInput -->|Empty/Esc| Cancel[Show: Update Cancelled]
    DescInput -->|New Desc| BuildPayload[Build PATCH Payload]
    Cancel --> ShowList
    
    BuildPayload --> ShowProgress[Display: Updating Gist...]
    ShowProgress --> CallAPI[PATCH to GitHub API]
    
    CallAPI --> APIResponse{Success?}
    APIResponse -->|No| Error[Show Error Message]
    APIResponse -->|Yes| UpdateLocal[Update Local Metadata]
    Error --> ShowList
    UpdateLocal --> Success[Show Success]
    Success --> ShowList
```

**State:**
- `gists []GistMetadata` - Available gists
- `selectedGist *GistMetadata` - Gist to update
- `newDescription string` - New description

**Actions:**
- PATCH request to update description only
- Update local metadata file
- Keep other gist data unchanged

---

## 12. Delete Gist Screen

```mermaid
flowchart TD
    Enter([Enter Delete Gist]) --> CheckToken{Token Available?}
    
    CheckToken -->|No| TokenError[Show: Token Required]
    TokenError --> Back1([Return to Gist Menu])
    
    CheckToken -->|Yes| LoadMetadata[Load Gist Metadata]
    LoadMetadata --> CheckGists{Gists Available?}
    
    CheckGists -->|No| NoGists[Show: No Gists to Delete]
    NoGists --> Back2([Return to Gist Menu])
    
    CheckGists -->|Yes| ShowList[Display Gist List]
    ShowList --> UserInput{User Input}
    
    UserInput -->|0/Esc| Back3([Return to Gist Menu])
    UserInput -->|Select| GetGist[Get Selected Gist]
    
    GetGist --> ShowWarning[Display Warning]
    ShowWarning --> ShowDetails[Show Gist Description]
    ShowDetails --> ConfirmPrompt[Prompt: Type 'yes' to confirm]
    
    ConfirmPrompt --> ConfirmInput{User Input}
    ConfirmInput -->|Not 'yes'| Cancel[Show: Deletion Cancelled]
    ConfirmInput -->|yes| ShowProgress[Display: Deleting Gist...]
    Cancel --> ShowList
    
    ShowProgress --> CallAPI[DELETE to GitHub API]
    CallAPI --> APIResponse{HTTP Status}
    
    APIResponse -->|204: Success| DeleteLocal[Delete Local Metadata]
    APIResponse -->|404: Not Found| DeleteLocal
    APIResponse -->|Other Error| ShowError[Show Error Details]
    
    DeleteLocal --> Success[Show Success]
    ShowError --> Fallback[Delete Local Anyway]
    Success --> ShowList
    Fallback --> ShowList
```

**State:**
- `gists []GistMetadata` - Available gists
- `selectedGist *GistMetadata` - Gist to delete
- `confirmed bool` - User confirmation

**Safety:**
- Requires explicit "yes" confirmation
- Shows gist details before delete
- Removes local metadata even on API error
- Handles 404 gracefully (already deleted)

---

## 13. Recover from Gist Screen

```mermaid
flowchart TD
    Enter([Enter Recover]) --> CheckLocal{Local Gists Exist?}
