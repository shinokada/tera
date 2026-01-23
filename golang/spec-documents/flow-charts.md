# TERA Screen Flow Charts

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
    QuickPlay --> SavePrompt: After playback
    SavePrompt --> MainMenu
```

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
    Playing -->|Stopped| Prompt[Show Save Prompt]
    Playing -->|q pressed| StopMPV[Stop MPV]
    StopMPV --> Prompt
    Prompt --> SaveChoice{User Choice}
    SaveChoice -->|1: Add to Favorites| AddFav[Add to My-favorites.json]
    SaveChoice -->|2: Return| Display
    AddFav --> Success[Show Success Message]
    Success --> Display
    
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
- Handle numeric shortcuts
- Navigate to selected screen

---

## 2. Play Screen

```mermaid
flowchart TD
    Enter([Enter Play Screen]) --> LoadLists[Load All Favorite Lists]
    LoadLists --> ShowLists[Display Lists with fzf-style]
    
    ShowLists --> ListInput{User Input}
    ListInput -->|Esc/Main Menu| Back([Return to Main Menu])
    ListInput -->|Select List| LoadStations[Load Stations from List]
    
    LoadStations --> ShowStations[Display Stations Sorted]
    ShowStations --> StationInput{User Input}
    
    StationInput -->|Esc/Main Menu| ShowLists
    StationInput -->|/| Filter[Enter Filter Mode]
    StationInput -->|Select| GetStation[Get Station Data]
    
    Filter --> FilterInput[Type Filter Text]
    FilterInput --> UpdateList[Filter Station List]
    UpdateList --> StationInput
    
    GetStation --> ShowInfo[Display Station Info]
    ShowInfo --> StartPlayer[Start MPV Player]
    StartPlayer --> Playing{Playback}
    
    Playing -->|q| Stop[Stop Playback]
    Playing -->|s| SaveToQuick[Save to Quick Favorites]
    Playing -->|Error| Error[Show Error Message]
    
    Stop --> Prompt[Show Save Prompt]
    Error --> Prompt
    SaveToQuick --> Continue[Continue Playing]
    Continue --> Playing
    
    Prompt --> PromptChoice{Save?}
    PromptChoice -->|Yes| AddQuick[Add to My-favorites.json]
    PromptChoice -->|No| ShowLists
    AddQuick --> CheckDupe{Already Exists?}
    CheckDupe -->|Yes| AlreadyMsg[Show Already Saved]
    CheckDupe -->|No| DoAdd[Add Station]
    DoAdd --> SuccessMsg[Show Success]
    AlreadyMsg --> ShowLists
    SuccessMsg --> ShowLists
```

**State:**
- `lists []string` - Available favorite lists
- `selectedList string` - Currently selected list
- `stations []Station` - Stations in selected list
- `filterText string` - Current filter
- `player *MPVPlayer` - Player instance

**Key Logic:**
- Stations displayed alphabetically (case-insensitive)
- Filter updates list in real-time
- Check for duplicates by StationUUID before adding

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
    
    CheckResults -->|Yes| ShowResults[Display Station List]
    
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
    AddToQuick --> Back2
    
    ShowSaveLists --> ListChoice{Select List}
    ListChoice -->|0| Submenu
    ListChoice -->|00| Back2
    ListChoice -->|Select| CheckDupe{Duplicate?}
    CheckDupe -->|Yes| DupeMsg[Show Already Exists]
    CheckDupe -->|No| SaveIt[Save to List]
    DupeMsg --> Back2
    SaveIt --> SuccessMsg[Show Success]
    SuccessMsg --> Back2
```

**State:**
- `results []Station` - Search results from API
- `filteredResults []Station` - After filter applied
- `selectedStation *Station` - Currently selected
- `filterText string` - Current filter

**Key Logic:**
- Check for duplicates by StationUUID
- Save prompt after playback ends
- Multiple navigation options (0, 00, Esc)

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
- \"My-favorites\" is protected (cannot delete/rename)
- Replace spaces with hyphens in names

---

## 6. Delete Station Screen

```mermaid
flowchart TD
    Enter([Enter Delete Station]) --> LoadLists[Load All Lists]
    LoadLists --> ShowLists[Display Lists]
    
    ShowLists --> ListInput{Select List}
    ListInput -->|Esc/0| Back([Return to Main Menu])
    ListInput -->|Select| LoadStations[Load Stations from List]
    
    LoadStations --> ShowStations[Display Stations]
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
    Playing -->|Error| PlayError[Show Error]
    
    Stop --> Prompt[Show Save Prompt]
    PlayError --> Prompt
    
    Prompt --> PromptChoice{Save to Favorites?}
    PromptChoice -->|Yes| AddToQuick[Add to My-favorites.json]
    PromptChoice -->|No| Back2([Return to Main Menu])
    AddToQuick --> CheckDupe{Already Exists?}
    CheckDupe -->|Yes| AlreadyMsg[Show Already Saved]
    CheckDupe -->|No| DoAdd[Add Station]
    AlreadyMsg --> Back2
    DoAdd --> Success[Show Success]
    Success --> Back2
```

**State:**
- `station *Station` - Random station selected
- `player *MPVPlayer` - Player instance

**Logic:**
- Query API for high-vote stations
- Select random from results
- Same save prompt flow as other play screens

---

## 8. Gist Menu Screen

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

## 9. Create Gist Screen

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

## 10. My Gists Screen

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

## 11. Token Management Screen

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

## 12. Update Gist Screen

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

## 13. Delete Gist Screen

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
- Requires explicit \"yes\" confirmation
- Shows gist details before delete
- Removes local metadata even on API error
- Handles 404 gracefully (already deleted)

---

## 14. Recover from Gist Screen

```mermaid
flowchart TD
    Enter([Enter Recover]) --> CheckLocal{Local Gists Exist?}
    
    CheckLocal -->|Yes| ShowSaved[Display Saved Gists List]
    CheckLocal -->|No| SkipList[Skip to Manual Input]
    
    ShowSaved --> ShowInstructions[Show: Can Select or Enter URL]
    ShowInstructions --> UserInput{User Input}
    SkipList --> ManualPrompt[Prompt: Enter Gist URL]
    
    UserInput -->|0/Esc| Back1([Return to Gist Menu])
    UserInput -->|Number| SelectSaved[Get Gist URL from List]
    UserInput -->|URL| UseURL[Use Provided URL]
    
    ManualPrompt --> ManualInput{Input}
    ManualInput -->|0/Esc| Back2([Return to Gist Menu])
    ManualInput -->|URL| UseURL
    
    SelectSaved --> ValidateURL[Validate URL Format]
    UseURL --> ValidateURL
    
    ValidateURL --> URLCheck{Valid?}
    URLCheck -->|No| InvalidURL[Show: Invalid URL Error]
    URLCheck -->|Yes| ShowCloning[Display: Cloning Gist...]
    InvalidURL --> Back3([Return to Gist Menu])
    
    ShowCloning --> GitClone[Execute: git clone]
    GitClone --> CloneResult{Success?}
    
    CloneResult -->|No| CloneError[Show Clone Error + Tips]
    CloneResult -->|Yes| FindJSON[Find .json Files in Clone]
    CloneError --> Back4([Return to Gist Menu])
    
    FindJSON --> CountFiles{JSON Files Found?}
    CountFiles -->|No| NoFiles[Show: No JSON Files]
    CountFiles -->|Yes| MoveFiles[Move Files to Favorites Dir]
    NoFiles --> Cleanup1[Remove Clone Directory]
    Cleanup1 --> Back5([Return to Gist Menu])
    
    MoveFiles --> Cleanup2[Remove Clone Directory]
    Cleanup2 --> CountMsg[Show: X Lists Downloaded]
    CountMsg --> Success[Display Success]
    Success --> Back6([Return to Gist Menu])
```

**State:**
- `savedGists []GistMetadata` - Local gist records
- `gistURL string` - URL to recover from
- `fileCount int` - Number of files recovered

**Actions:**
- git clone to temp directory
- Find all .json files
- Move to ~/.config/tera/favorite/
- Clean up clone directory
- Show count of files recovered

---

## Error Handling Patterns

```mermaid
flowchart TD
    Error([Error Occurs]) --> Classify{Error Type}
    
    Classify -->|Network| NetworkError[Network Error Handler]`
    Classify -->|API| APIError[API Error Handler]
    Classify -->|File| FileError[File Error Handler]
    Classify -->|Player| PlayerError[Player Error Handler]
    Classify -->|Validation| ValidationError[Validation Error Handler]
    
    NetworkError --> ShowNetMsg[Display: Network Issue]
    ShowNetMsg --> NetTips[Show Connection Tips]
    NetTips --> OfferRetry{Retry Available?}
    
    APIError --> ShowAPIMsg[Display: API Error]
    ShowAPIMsg --> APITips[Show API-Specific Tips]
    APITips --> OfferRetry
    
    FileError --> ShowFileMsg[Display: File Operation Failed]
    ShowFileMsg --> FileTips[Show File Permission Tips]
    FileTips --> OfferRetry
    
    PlayerError --> ShowPlayerMsg[Display: Playback Error]
    ShowPlayerMsg --> PlayerTips[Show MPV Installation Tips]
    PlayerTips --> OfferRetry
    
    ValidationError --> ShowValMsg[Display: Validation Failed]
    ShowValMsg --> ValTips[Show Input Format Tips]
    ValTips --> ReturnToInput[Return to Input Screen]
    
    OfferRetry -->|Yes| RetryPrompt[Show Retry Option]
    OfferRetry -->|No| WaitReturn[Wait for Enter to Continue]
    
    RetryPrompt --> UserChoice{User Choice}
    UserChoice -->|Retry| RetryAction[Retry Original Action]
    UserChoice -->|Cancel| WaitReturn
    
    RetryAction --> Success{Succeeds?}
    Success -->|Yes| ContinueFlow[Continue Normal Flow]
    Success -->|No| Error
    
    WaitReturn --> PreviousScreen[Return to Previous Screen]
```

---

## State Transitions

```mermaid
stateDiagram-v2
    [*] --> Initializing
    Initializing --> MainMenu: Config Loaded
    
    MainMenu --> PlayScreen: User selects Play
    MainMenu --> SearchMenu: User selects Search
    MainMenu --> ListMenu: User selects List
    MainMenu --> DeleteStation: User selects Delete
    MainMenu --> LuckyScreen: User selects Lucky
    MainMenu --> GistMenu: User selects Gist
    MainMenu --> QuickPlay: User selects 10-19
    MainMenu --> Exiting: User quits
    
    PlayScreen --> StationPlaying: Station selected
    SearchMenu --> SearchResults: Search executed
    ListMenu --> ListOperation: Operation selected
    DeleteStation --> ConfirmDelete: Station selected
    LuckyScreen --> StationPlaying: Random station found
    GistMenu --> GistOperation: Gist action selected
    QuickPlay --> StationPlaying: Quick favorite selected
    
    StationPlaying --> SavePrompt: Playback ends
    SearchResults --> StationPlaying: Play selected
    SearchResults --> SaveStation: Save selected
    
    SavePrompt --> MainMenu: User responds
    SaveStation --> SearchResults: Save complete
    ListOperation --> ListMenu: Operation complete
    ConfirmDelete --> DeleteStation: Delete complete
    GistOperation --> GistMenu: Operation complete
    
    PlayScreen --> MainMenu: Back pressed
    SearchMenu --> MainMenu: Back pressed
    SearchResults --> SearchMenu: Back pressed
    ListMenu --> MainMenu: Back pressed
    DeleteStation --> MainMenu: Back pressed
    GistMenu --> MainMenu: Back pressed
    
    Exiting --> [*]: Cleanup complete
```

---

## Data Flow

```mermaid
flowchart LR
    subgraph External
        API[Radio Browser API]
        GitHub[GitHub Gist API]
        MPV[MPV Player Process]
    end
    
    subgraph Storage
        Config[Config File]
        Favorites[Favorite Lists JSON]
        Metadata[Gist Metadata JSON]
        Token[Token Storage]
    end
    
    subgraph Application
        UI[Bubble Tea UI]
        APIClient[API Client]
        GistClient[Gist Client]
        Storage[Storage Layer]
        Player[Player Controller]
    end
    
    UI -->|Search Query| APIClient
    APIClient -->|HTTP POST| API
    API -->|Stations JSON| APIClient
    APIClient -->|Station Data| UI
    
    UI -->|Play Command| Player
    Player -->|Start Process| MPV
    MPV -->|Stream Audio| Player
    Player -->|Status| UI
    
    UI -->|Load Lists| Storage
    Storage -->|Read| Favorites
    Favorites -->|JSON Data| Storage
    Storage -->|Station Lists| UI
    
    UI -->|Save Station| Storage
    Storage -->|Write| Favorites
    
    UI -->|Create Gist| GistClient
    GistClient -->|Load Token| Token
    GistClient -->|Read Lists| Favorites
    GistClient -->|POST| GitHub
    GitHub -->|Gist URL| GistClient
    GistClient -->|Save| Metadata
    GistClient -->|Result| UI
    
    UI -->|Load Config| Storage
    Storage -->|Read| Config
    Config -->|Settings| Storage
    Storage -->|Config Data| UI
```
