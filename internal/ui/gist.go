package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/gist"
	"github.com/shinokada/tera/v3/internal/storage"
	"github.com/shinokada/tera/v3/internal/ui/components"
)

// ── States ────────────────────────────────────────────────────────────────────

type gistState int

const (
	gistStateMenu             gistState = iota
	gistStateCreateVisibility           // public/secret choice
	gistStateCreateName                 // gist description input
	gistStateCreate                     // transient: uploading
	gistStateList                       // browse my gists
	gistStateUpdate                     // pick gist to update
	gistStateDelete                     // pick gist to delete
	gistStateRecover                    // pick gist to recover
	gistStateImportURL                  // import from URL
	gistStateTokenMenu
	gistStateTokenSetup
	gistStateTokenView
	gistStateTokenDelete
	gistStateUpdateInput
	gistStateDeleteConfirm
	// ── Phase 3: Sync & Backup ────────────────────────────────────────────────
	gistStateExportChecklist      // category checklist → zip export
	gistStateExportPath           // path prompt for zip destination
	gistStateRestoreZipPath       // path prompt for zip source
	gistStateRestoreZipChecklist  // category checklist for zip restore
	gistStateSyncGistChecklist    // category checklist → Gist push
	gistStateRestoreGistURL       // URL input for Gist restore
	gistStateRestoreGistChecklist // category checklist for Gist pull
	gistStateOverwriteWarn        // warn before clobbering existing files
	gistStateSyncProgress         // transient: show result then return to menu
)

// overwriteSource distinguishes which restore flow triggered the overwrite warning.
type overwriteSource int

const (
	overwriteSourceZip  overwriteSource = iota
	overwriteSourceGist
)

// ── Internal message types ─────────────────────────────────────────────────────

// Returned by async Cmds to signal multi-step flow transitions.
type zipInspectedMsg struct {
	path      string
	available storage.SyncPrefs
}
type zipConflictCheckMsg struct {
	zipPath   string
	prefs     storage.SyncPrefs
	conflicts []string
}
type gistRestoreAvailableMsg struct {
	g         *gist.Gist
	available storage.SyncPrefs
}
type gistConflictCheckMsg struct {
	g         *gist.Gist
	prefs     storage.SyncPrefs
	conflicts []string
}

// ── Existing message types (kept for compatibility) ────────────────────────────

type successMsg struct{ msg string }
type gistsMsg []*gist.GistMetadata
type tokenSavedMsg struct{ token, user string }
type tokenDeletedMsg struct{}

// ── gistItem (list.Item) ──────────────────────────────────────────────────────

type gistItem struct{ meta *gist.GistMetadata }

func (i gistItem) Title() string       { return i.meta.Description }
func (i gistItem) Description() string { return i.meta.CreatedAt.Format("2006-01-02 15:04") }
func (i gistItem) FilterValue() string { return i.meta.Description }

// ── GistModel ─────────────────────────────────────────────────────────────────

type GistModel struct {
	// core
	state        gistState
	favoritePath string
	gistClient   *gist.Client
	// list widgets
	menuList       list.Model
	tokenMenuList  list.Model
	visibilityMenu list.Model
	gistList       list.Model
	// gist data
	gists           []*gist.GistMetadata
	selectedGist    *gist.GistMetadata
	gistDescription string
	createPublic    bool
	// text input
	textInput    textinput.Model
	// Phase 3: backup/sync
	checklist      components.ChecklistModel
	syncPrefs      storage.SyncPrefs
	backupManager  *storage.BackupManager
	gistSyncMgr    *storage.GistSyncManager
	pendingZipPath string          // zip path set after path prompt
	pendingPrefs   storage.SyncPrefs // prefs confirmed before conflict check
	overwritePaths []string          // files that would be clobbered
	overwriteSrc   overwriteSource
	pendingGist    *gist.Gist // gist fetched during restore-from-URL flow
	// ui
	message        string
	messageIsError bool
	width          int
	height         int
	token          string
	quitting       bool
}

// ── Constructor ───────────────────────────────────────────────────────────────

func NewGistModel(favoritePath string) GistModel {
	menuItems := []components.MenuItem{
		// — Favorites Gist —
		components.NewMenuItem("Create favorites gist", "Upload favorites to a new secret gist", "1"),
		components.NewMenuItem("My favorites gists", "View and manage your saved gists", "2"),
		components.NewMenuItem("Recover favorites gist", "Download and restore favorites from a gist", "3"),
		components.NewMenuItem("Import favorites from URL", "Import favorites from any public gist URL", "4"),
		components.NewMenuItem("Update favorites gist", "Update description of an existing gist", "5"),
		components.NewMenuItem("Delete favorites gist", "Remove a gist permanently", "6"),
		// — Full Backup —
		components.NewMenuItem("Export backup (zip)", "Save all data to a local zip file", "7"),
		components.NewMenuItem("Restore from backup (zip)", "Restore data from a local zip file", "8"),
		components.NewMenuItem("Sync all data to Gist", "Push selected data to a backup Gist", "9"),
		components.NewMenuItem("Restore all data from Gist", "Pull selected data from a backup Gist", "a"),
		// — Account —
		components.NewMenuItem("Token Management", "Manage your GitHub Personal Access Token", "t"),
	}
	menuList := components.CreateMenu(menuItems, "Sync & Backup", 0, 0)

	tokenMenuItems := []components.MenuItem{
		components.NewMenuItem("Setup/Change Token", "Configure your GitHub Personal Access Token", "1"),
		components.NewMenuItem("View Current Token", "See your masked token", "2"),
		components.NewMenuItem("Validate Token", "Test if your token is valid", "3"),
		components.NewMenuItem("Delete Token", "Remove your stored token", "4"),
	}
	tokenMenuList := components.CreateMenu(tokenMenuItems, "Token Menu", 0, 0)

	visibilityMenuItems := []components.MenuItem{
		components.NewMenuItem("Secret gist", "Only you can see this gist (recommended)", "1"),
		components.NewMenuItem("Public gist", "Anyone can see this gist", "2"),
	}
	visibilityMenu := components.CreateMenu(visibilityMenuItems, "Visibility", 0, 0)

	delegate := createStyledDelegate()
	gistList := list.New([]list.Item{}, delegate, 50, 20)
	gistList.Title = "My Gists"
	gistList.SetShowStatusBar(false)
	gistList.SetShowHelp(false)
	gistList.SetShowTitle(false)
	gistList.SetShowPagination(false)

	ti := textinput.New()
	ti.Placeholder = "Type here..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	token, tokenErr := gist.LoadToken()
	syncPrefs, _ := storage.LoadSyncPrefs()

	backupMgr, backupErr := storage.NewBackupManager()

	m := GistModel{
		state:          gistStateMenu,
		favoritePath:   favoritePath,
		menuList:       menuList,
		tokenMenuList:  tokenMenuList,
		visibilityMenu: visibilityMenu,
		gistList:       gistList,
		textInput:      ti,
		token:          token,
		syncPrefs:      syncPrefs,
		backupManager:  backupMgr,
	}

	if backupErr != nil {
		m.message = fmt.Sprintf("Warning: backup features unavailable: %v", backupErr)
		m.messageIsError = true
	} else if tokenErr != nil {
		m.message = fmt.Sprintf("Warning: could not load token: %v", tokenErr)
		m.messageIsError = true
	}

	if token != "" {
		m.gistClient = gist.NewClient(token)
		if sm, err := storage.NewGistSyncManager(m.gistClient); err == nil {
			m.gistSyncMgr = sm
		} else if m.message == "" {
			// Only set if no higher-priority warning is already shown.
			m.message = fmt.Sprintf("Warning: Gist sync unavailable: %v", err)
			m.messageIsError = true
		}
	}

	return m
}

func (m GistModel) Init() tea.Cmd { return nil }

// ── Update ────────────────────────────────────────────────────────────────────

func (m GistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// ── global messages ────────────────────────────────────────────────────────
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		// Global Esc handling per state.
		switch m.state {
		case gistStateMenu:
			if msg.String() == "esc" || msg.String() == "0" {
				m.quitting = true
				return m, func() tea.Msg { return backToMainMsg{} }
			}
		case gistStateCreateVisibility,
			gistStateList, gistStateUpdate, gistStateDelete,
			gistStateRecover, gistStateTokenMenu:
			if msg.String() == "esc" {
				m.state = gistStateMenu
				m.message = ""
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menuList.SetWidth(msg.Width)
		m.tokenMenuList.SetWidth(msg.Width)
		m.visibilityMenu.SetWidth(msg.Width)
		m.gistList.SetWidth(msg.Width)
		m.gistList.SetHeight(availableListHeight(msg.Height))
		m.checklist.SetWidth(msg.Width)
		return m, nil

	case errMsg:
		m.message = msg.err.Error()
		m.messageIsError = true
		// Dismiss transient states on error, returning to the menu.
		// gistStateRestoreGistURL is intentionally excluded: errors received
		// while fetching stay on the URL form so the user can correct a typo
		// or retry without losing their input.
		switch m.state {
		case gistStateCreate, gistStateList, gistStateUpdate, gistStateDelete,
			gistStateRecover, gistStateImportURL, gistStateSyncProgress,
			gistStateExportPath, gistStateRestoreZipPath,
			gistStateRestoreZipChecklist, gistStateSyncGistChecklist,
			gistStateRestoreGistChecklist:
			m.state = gistStateMenu
		}
		return m, nil

	case successMsg:
		m.message = msg.msg
		m.messageIsError = false
		switch m.state {
		case gistStateCreate, gistStateRecover, gistStateUpdateInput,
			gistStateDeleteConfirm, gistStateImportURL, gistStateSyncProgress,
			gistStateExportPath, gistStateSyncGistChecklist:
			m.state = gistStateMenu
		}
		return m, nil

	case gistsMsg:
		m.gists = msg
		items := make([]list.Item, len(m.gists))
		for i, g := range m.gists {
			items[i] = gistItem{meta: g}
		}
		m.gistList.SetItems(items)
		if m.height > 0 {
			m.gistList.SetHeight(availableListHeight(m.height))
		} else {
			m.gistList.SetHeight(10)
		}
		return m, nil

	case tokenSavedMsg:
		m.token = msg.token
		m.gistClient = gist.NewClient(m.token)
		if sm, err := storage.NewGistSyncManager(m.gistClient); err == nil {
			m.gistSyncMgr = sm
			m.message = fmt.Sprintf("Token saved! User: %s", msg.user)
		} else {
			m.gistSyncMgr = nil
			m.message = fmt.Sprintf("Token saved (User: %s) but Gist sync unavailable: %v", msg.user, err)
		}
		m.messageIsError = false
		m.state = gistStateTokenMenu
		return m, nil

	case tokenDeletedMsg:
		m.token = ""
		m.gistClient = nil
		m.gistSyncMgr = nil
		m.message = "✓ Token has been deleted successfully!"
		m.messageIsError = false
		m.state = gistStateTokenMenu
		return m, nil

	// ── Phase 3: checklist events ──────────────────────────────────────────────
	case components.ChecklistConfirmedMsg:
		return m.handleChecklistConfirmed(msg)

	case components.ChecklistCancelledMsg:
		m.state = gistStateMenu
		m.message = ""
		return m, nil

	// ── Phase 3: intermediate async results ────────────────────────────────────

	case zipInspectedMsg:
		// Zip opened OK — show checklist of available categories.
		m.pendingZipPath = msg.path
		m.checklist = availableChecklist("Select categories to restore from zip:", msg.available)
		m.state = gistStateRestoreZipChecklist
		return m, nil

	case zipConflictCheckMsg:
		if len(msg.conflicts) == 0 {
			// No conflicts — restore immediately (force=false is fine).
			m.state = gistStateSyncProgress
			return m, m.restoreZipCmd(msg.zipPath, msg.prefs, false)
		}
		// Conflicts exist — warn user.
		m.pendingZipPath = msg.zipPath
		m.pendingPrefs = msg.prefs
		m.overwritePaths = msg.conflicts
		m.overwriteSrc = overwriteSourceZip
		m.state = gistStateOverwriteWarn
		return m, nil

	case gistRestoreAvailableMsg:
		m.pendingGist = msg.g
		m.checklist = availableChecklist("Select categories to restore from Gist:", msg.available)
		m.state = gistStateRestoreGistChecklist
		return m, nil

	case gistConflictCheckMsg:
		if len(msg.conflicts) == 0 {
			m.state = gistStateSyncProgress
			return m, m.restoreGistCmd(msg.g, msg.prefs, false)
		}
		m.pendingGist = msg.g
		m.pendingPrefs = msg.prefs
		m.overwritePaths = msg.conflicts
		m.overwriteSrc = overwriteSourceGist
		m.state = gistStateOverwriteWarn
		return m, nil
	}

	// ── State machine ──────────────────────────────────────────────────────────
	switch m.state {
	case gistStateMenu:
		return m.updateMenu(msg)

	case gistStateCreateVisibility:
		return m.updateCreateVisibility(msg)

	case gistStateCreateName:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				m.gistDescription = m.textInput.Value()
				if m.gistDescription == "" {
					m.gistDescription = fmt.Sprintf("TERA Radio Favorites - %s", time.Now().Format("2006-01-02 15:04:05"))
				}
				m.state = gistStateCreate
				return m.initCreateGist()
			case "esc":
				m.state = gistStateCreateVisibility
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case gistStateCreate:
		return m, nil

	case gistStateList, gistStateUpdate, gistStateDelete, gistStateRecover:
		return m.updateGistList(msg)

	case gistStateUpdateInput:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				return m, m.updateGistCmd(m.selectedGist.ID, m.textInput.Value())
			case "esc":
				m.state = gistStateUpdate
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

	case gistStateDeleteConfirm:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				if strings.ToLower(m.textInput.Value()) == "yes" {
					return m, m.deleteGistCmd(m.selectedGist.ID)
				}
				m.message = "Deletion cancelled"
				m.messageIsError = true
				m.state = gistStateMenu
				return m, nil
			case "esc":
				m.state = gistStateDelete
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

	case gistStateImportURL:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				input := m.textInput.Value()
				gistID, err := gist.ParseGistURL(input)
				if err != nil {
					m.message = fmt.Sprintf("Invalid URL or ID: %v", err)
					m.messageIsError = true
					return m, nil
				}
				return m, m.importGistCmd(gistID)
			case "esc":
				m.state = gistStateMenu
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

	case gistStateTokenMenu:
		return m.updateTokenMenu(msg)

	case gistStateTokenSetup:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				if token := m.textInput.Value(); token != "" {
					return m, m.saveTokenCmd(token)
				}
			case "esc":
				m.state = gistStateTokenMenu
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

	case gistStateTokenView:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" || keyMsg.String() == "esc" {
				m.state = gistStateTokenMenu
				return m, nil
			}
		}

	case gistStateTokenDelete:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				if strings.ToLower(m.textInput.Value()) == "yes" {
					return m, m.deleteTokenCmd()
				}
				m.message = "Token deletion cancelled"
				m.messageIsError = false
				m.state = gistStateTokenMenu
				return m, nil
			case "esc":
				m.state = gistStateTokenMenu
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

	case gistStateRestoreGistURL:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				input := m.textInput.Value()
				gistID, err := gist.ParseGistURL(input)
				if err != nil {
					m.message = fmt.Sprintf("Invalid URL or ID: %v", err)
					m.messageIsError = true
					return m, nil
				}
				m.message = "Fetching Gist\u2026"
				m.messageIsError = false
				return m, m.doFetchAvailableGistCategoriesCmd(gistID)
			case "esc":
				m.state = gistStateMenu
				return m, nil
			}
		}
		var urlCmd tea.Cmd
		m.textInput, urlCmd = m.textInput.Update(msg)
		return m, urlCmd

	// ── Phase 3: checklist states forward to the checklist widget ─────────────
	case gistStateExportChecklist, gistStateRestoreZipChecklist,
		gistStateSyncGistChecklist, gistStateRestoreGistChecklist:
		var checkCmd tea.Cmd
		m.checklist, checkCmd = m.checklist.Update(msg)
		cmds = append(cmds, checkCmd)

	// ── Phase 3: path prompts ─────────────────────────────────────────────────
	case gistStateExportPath:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				m.state = gistStateSyncProgress
				return m, m.doExportZipCmd(m.textInput.Value(), m.pendingPrefs)
			case "esc":
				m.state = gistStateExportChecklist
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

	case gistStateRestoreZipPath:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				return m, m.doInspectZipCmd(m.textInput.Value())
			case "esc":
				m.state = gistStateMenu
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

	// ── Phase 3: overwrite confirmation ───────────────────────────────────────
	case gistStateOverwriteWarn:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				m.state = gistStateSyncProgress
				switch m.overwriteSrc {
				case overwriteSourceZip:
					return m, m.restoreZipCmd(m.pendingZipPath, m.pendingPrefs, true)
				case overwriteSourceGist:
					return m, m.restoreGistCmd(m.pendingGist, m.pendingPrefs, true)
				}
			case "esc":
				m.state = gistStateMenu
				m.message = "Restore cancelled."
				m.messageIsError = false
				return m, nil
			}
		}

	case gistStateSyncProgress:
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// ── Sub-update helpers ────────────────────────────────────────────────────────

func (m GistModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	var selected int
	m.menuList, selected = components.HandleMenuKey(keyMsg, m.menuList)
	if selected < 0 {
		return m, nil
	}
	m.message = ""
	m.messageIsError = false

	switch selected {
	case 0: // Create favorites gist
		if m.token == "" {
			m.message = "No token configured. Set up a token first."
			m.messageIsError = true
			return m, nil
		}
		m.state = gistStateCreateVisibility
		return m, nil

	case 1: // My favorites gists
		m.state = gistStateList
		return m.initGistList()

	case 2: // Recover favorites gist
		m.state = gistStateRecover
		return m.initGistList()

	case 3: // Import from URL
		m.state = gistStateImportURL
		m.textInput.Placeholder = "https://gist.github.com/user/id or gist ID"
		m.textInput.SetValue("")
		m.textInput.EchoMode = textinput.EchoNormal
		m.textInput.Focus()
		return m, nil

	case 4: // Update favorites gist
		m.state = gistStateUpdate
		return m.initGistList()

	case 5: // Delete favorites gist
		m.state = gistStateDelete
		return m.initGistList()

	case 6: // Export backup (zip)
		m.checklist = m.buildDefaultChecklist("Select categories to include in the backup zip:")
		m.state = gistStateExportChecklist
		return m, nil

	case 7: // Restore from backup (zip)
		m.state = gistStateRestoreZipPath
		m.textInput.Placeholder = "Path to backup zip file"
		defaultPath, _ := storage.DefaultBackupPath()
		m.textInput.SetValue(defaultPath)
		m.textInput.EchoMode = textinput.EchoNormal
		m.textInput.Focus()
		return m, nil

	case 8: // Sync all data to Gist
		if m.token == "" {
			m.message = "No token configured. Set up a token first."
			m.messageIsError = true
			return m, nil
		}
		if m.gistSyncMgr == nil {
			m.message = "Gist sync unavailable — check token setup."
			m.messageIsError = true
			return m, nil
		}
		m.checklist = m.buildDefaultChecklist("Select categories to sync to Gist:")
		m.state = gistStateSyncGistChecklist
		return m, nil

	case 9: // Restore all data from Gist
		m.state = gistStateRestoreGistURL
		m.textInput.Placeholder = "https://gist.github.com/user/id or gist ID"
		m.textInput.SetValue("")
		m.textInput.EchoMode = textinput.EchoNormal
		m.textInput.Focus()
		return m, nil

	case 10: // Token Management
		m.state = gistStateTokenMenu
		return m.initTokenMenu()
	}

	return m, nil
}

func (m GistModel) updateCreateVisibility(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if keyMsg.String() == "esc" {
		m.state = gistStateMenu
		return m, nil
	}
	var selected int
	m.visibilityMenu, selected = components.HandleMenuKey(keyMsg, m.visibilityMenu)
	if selected < 0 {
		return m, nil
	}
	m.createPublic = selected == 1
	m.state = gistStateCreateName
	m.textInput.Placeholder = "Enter gist name (or press Enter for default)"
	m.textInput.SetValue(fmt.Sprintf("TERA Radio Favorites - %s", time.Now().Format("2006-01-02 15:04:05")))
	m.textInput.EchoMode = textinput.EchoNormal
	m.textInput.Focus()
	return m, nil
}

func (m GistModel) updateGistList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "enter" {
			if selectedItem, ok := m.gistList.SelectedItem().(gistItem); ok {
				m.selectedGist = selectedItem.meta
				switch m.state {
				case gistStateList:
					if err := openBrowser(m.selectedGist.URL); err != nil {
						m.message = fmt.Sprintf("Failed to open browser: %v", err)
						m.messageIsError = true
					} else {
						m.message = fmt.Sprintf("Opened %s in browser", m.selectedGist.URL)
					}
					return m, nil
				case gistStateUpdate:
					m.textInput.Placeholder = "New description"
					m.textInput.SetValue(m.selectedGist.Description)
					m.textInput.EchoMode = textinput.EchoNormal
					m.state = gistStateUpdateInput
					m.textInput.Focus()
					return m, nil
				case gistStateDelete:
					m.textInput.Placeholder = "Type 'yes' to confirm"
					m.textInput.SetValue("")
					m.textInput.EchoMode = textinput.EchoNormal
					m.state = gistStateDeleteConfirm
					m.textInput.Focus()
					return m, nil
				case gistStateRecover:
					return m, m.recoverGistCmd(m.selectedGist.ID)
				}
			}
		}
	}
	m.gistList, cmd = m.gistList.Update(msg)
	return m, cmd
}

func (m GistModel) updateTokenMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if keyMsg.String() == "esc" {
		m.state = gistStateMenu
		m.message = ""
		return m, nil
	}
	var selected int
	m.tokenMenuList, selected = components.HandleMenuKey(keyMsg, m.tokenMenuList)
	if selected < 0 {
		return m, nil
	}
	m.message = ""
	m.messageIsError = false
	switch selected {
	case 0:
		m.state = gistStateTokenSetup
		m.textInput.Placeholder = "Paste GitHub Token"
		m.textInput.SetValue("")
		m.textInput.EchoMode = textinput.EchoPassword
		m.textInput.Focus()
	case 1:
		m.state = gistStateTokenView
	case 2:
		if m.token == "" {
			m.message = "No token configured!"
			m.messageIsError = true
			return m, nil
		}
		return m, m.validateTokenCmd()
	case 3:
		if m.token == "" {
			m.message = "No token to delete!"
			m.messageIsError = true
			return m, nil
		}
		m.state = gistStateTokenDelete
		m.textInput.Placeholder = "Type 'yes' to confirm"
		m.textInput.SetValue("")
		m.textInput.EchoMode = textinput.EchoNormal
		m.textInput.Focus()
	}
	return m, nil
}

// ── Checklist helpers ─────────────────────────────────────────────────────────

// buildDefaultChecklist returns a ChecklistModel pre-populated from syncPrefs.
func (m GistModel) buildDefaultChecklist(title string) components.ChecklistModel {
	items := []components.ChecklistItem{
		{Key: "favorites", Label: "Favorites (playlists)", Checked: m.syncPrefs.Favorites},
		{Key: "settings", Label: "Settings (config.yaml)", Detail: "machine-specific", Checked: m.syncPrefs.Settings},
		{Key: "ratings_votes", Label: "Ratings & votes", Checked: m.syncPrefs.RatingsVotes},
		{Key: "blocklist", Label: "Blocklist", Checked: m.syncPrefs.Blocklist},
		{Key: "metadata_tags", Label: "Station metadata & tags", Checked: m.syncPrefs.MetadataTags},
		{Key: "search_history", Label: "Search history", Checked: m.syncPrefs.SearchHistory},
	}
	return components.NewChecklistModel(title, items)
}

// availableChecklist builds a checklist from a SyncPrefs mask (only truthy
// categories shown, all pre-checked). Used for restore flows.
func availableChecklist(title string, available storage.SyncPrefs) components.ChecklistModel {
	type entry struct {
		key   string
		label string
		avail bool
	}
	all := []entry{
		{"favorites", "Favorites (playlists)", available.Favorites},
		{"settings", "Settings (config.yaml)", available.Settings},
		{"ratings_votes", "Ratings & votes", available.RatingsVotes},
		{"blocklist", "Blocklist", available.Blocklist},
		{"metadata_tags", "Station metadata & tags", available.MetadataTags},
		{"search_history", "Search history", available.SearchHistory},
	}
	var items []components.ChecklistItem
	for _, e := range all {
		if e.avail {
			items = append(items, components.ChecklistItem{Key: e.key, Label: e.label, Checked: true})
		}
	}
	return components.NewChecklistModel(title, items)
}

// checklistToPrefs converts confirmed checklist items back to SyncPrefs.
func checklistToPrefs(items []components.ChecklistItem) storage.SyncPrefs {
	var p storage.SyncPrefs
	for _, item := range items {
		switch item.Key {
		case "favorites":
			p.Favorites = item.Checked
		case "settings":
			p.Settings = item.Checked
		case "ratings_votes":
			p.RatingsVotes = item.Checked
		case "blocklist":
			p.Blocklist = item.Checked
		case "metadata_tags":
			p.MetadataTags = item.Checked
		case "search_history":
			p.SearchHistory = item.Checked
		}
	}
	return p
}

// handleChecklistConfirmed dispatches based on the current checklist state.
func (m GistModel) handleChecklistConfirmed(msg components.ChecklistConfirmedMsg) (tea.Model, tea.Cmd) {
	prefs := checklistToPrefs(msg.Items)
	if prefs == (storage.SyncPrefs{}) {
		m.message = "Select at least one category."
		m.messageIsError = true
		return m, nil // stay on the checklist
	}

	switch m.state {
	case gistStateExportChecklist:
		// Persist selections — this is a write/export flow, not a restore.
		m.syncPrefs = prefs
		if err := storage.SaveSyncPrefs(prefs); err != nil {
			m.message = fmt.Sprintf("Warning: could not save category preferences: %v", err)
			m.messageIsError = true
		}
		// Move to path prompt
		defaultPath, _ := storage.DefaultBackupPath()
		m.textInput.Placeholder = "Save location"
		m.textInput.SetValue(defaultPath)
		m.textInput.EchoMode = textinput.EchoNormal
		m.textInput.Focus()
		m.pendingPrefs = prefs
		m.state = gistStateExportPath
		return m, nil

	case gistStateRestoreZipChecklist:
		m.pendingPrefs = prefs
		return m, m.doCheckZipConflictsCmd(m.pendingZipPath, prefs)

	case gistStateSyncGistChecklist:
		m.syncPrefs = prefs
		if err := storage.SaveSyncPrefs(prefs); err != nil {
			m.message = fmt.Sprintf("Warning: could not save category preferences: %v", err)
			m.messageIsError = true
		}
		m.pendingPrefs = prefs
		m.state = gistStateSyncProgress
		return m, m.doSyncToGistCmd(prefs)

	case gistStateRestoreGistChecklist:
		m.pendingPrefs = prefs
		return m, m.doCheckGistConflictsCmd(m.pendingGist, prefs)
	}

	return m, nil
}

// ── Async Cmd factories (Phase 3) ─────────────────────────────────────────────

func (m GistModel) doExportZipCmd(rawPath string, prefs storage.SyncPrefs) tea.Cmd {
	bm := m.backupManager
	return func() tea.Msg {
		if bm == nil {
			return errMsg{fmt.Errorf("backup manager unavailable")}
		}
		resolved, err := storage.ResolveBackupPath(rawPath)
		if err != nil {
			return errMsg{fmt.Errorf("invalid path: %w", err)}
		}
		if err := bm.Export(resolved, prefs); err != nil {
			return errMsg{fmt.Errorf("export failed: %w", err)}
		}
		return successMsg{fmt.Sprintf("✓ Backup saved to %s", resolved)}
	}
}

func (m GistModel) doInspectZipCmd(rawPath string) tea.Cmd {
	bm := m.backupManager
	return func() tea.Msg {
		if bm == nil {
			return errMsg{fmt.Errorf("backup manager unavailable")}
		}
		resolved, err := storage.ResolveBackupPath(rawPath)
		if err != nil {
			return errMsg{fmt.Errorf("invalid path: %w", err)}
		}
		available, err := bm.ListArchiveCategories(resolved)
		if err != nil {
			return errMsg{fmt.Errorf("cannot read zip: %w", err)}
		}
		if available == (storage.SyncPrefs{}) {
			return errMsg{fmt.Errorf("zip contains no recognised tera data files")}
		}
		return zipInspectedMsg{path: resolved, available: available}
	}
}

func (m GistModel) doCheckZipConflictsCmd(zipPath string, prefs storage.SyncPrefs) tea.Cmd {
	bm := m.backupManager
	return func() tea.Msg {
		if bm == nil {
			return errMsg{fmt.Errorf("backup manager unavailable")}
		}
		conflicts, err := bm.ConflictingFiles(zipPath, prefs)
		if err != nil {
			return errMsg{fmt.Errorf("conflict check failed: %w", err)}
		}
		return zipConflictCheckMsg{zipPath: zipPath, prefs: prefs, conflicts: conflicts}
	}
}

func (m GistModel) restoreZipCmd(zipPath string, prefs storage.SyncPrefs, force bool) tea.Cmd {
	bm := m.backupManager
	return func() tea.Msg {
		if bm == nil {
			return errMsg{fmt.Errorf("backup manager unavailable")}
		}
		if err := bm.Restore(zipPath, prefs, force); err != nil {
			return errMsg{fmt.Errorf("restore failed: %w", err)}
		}
		return successMsg{"✓ Data restored successfully from zip."}
	}
}

func (m GistModel) doFetchAvailableGistCategoriesCmd(gistID string) tea.Cmd {
	mgr := m.gistSyncMgr
	authClient := m.gistClient
	return func() tea.Msg {
		// Try authenticated fetch first (works for both public and private gists).
		var g *gist.Gist
		var fetchErr error
		if authClient != nil {
			g, fetchErr = authClient.GetGist(gistID)
		}
		// Fall back to unauthenticated fetch for public gists.
		if g == nil || fetchErr != nil {
			g, fetchErr = gist.GetGistPublic(gistID)
			if fetchErr != nil {
				if authClient == nil {
					return errMsg{fmt.Errorf("failed to fetch Gist: %w (if private, set up a token first)", fetchErr)}
				}
				return errMsg{fmt.Errorf("failed to fetch Gist: %w", fetchErr)}
			}
		}
		// Determine which categories the gist contains.
		var available storage.SyncPrefs
		var err error
		if mgr != nil {
			available, err = mgr.AvailableCategoriesFromGist(g)
			if err != nil {
				return errMsg{fmt.Errorf("failed to inspect Gist: %w", err)}
			}
		} else {
			available = storage.AvailableCategoriesFromGistFiles(g.Files)
		}
		if available == (storage.SyncPrefs{}) {
			return errMsg{fmt.Errorf("no recognisable tera data found in this Gist")}
		}
		return gistRestoreAvailableMsg{g: g, available: available}
	}
}

func (m GistModel) doSyncToGistCmd(prefs storage.SyncPrefs) tea.Cmd {
	mgr := m.gistSyncMgr
	return func() tea.Msg {
		if mgr == nil {
			return errMsg{fmt.Errorf("no Gist sync manager — token required")}
		}
		if err := mgr.Push(prefs); err != nil {
			return errMsg{fmt.Errorf("gist sync failed: %w", err)}
		}
		return successMsg{fmt.Sprintf("✓ Data synced to Gist (%s).", storage.BackupGistDescription)}
	}
}

func (m GistModel) doCheckGistConflictsCmd(g *gist.Gist, prefs storage.SyncPrefs) tea.Cmd {
	mgr := m.gistSyncMgr
	return func() tea.Msg {
		var conflicts []string
		var err error
		if mgr != nil {
			conflicts, err = mgr.ConflictingGistFiles(g, prefs)
		} else {
			// No token — use standalone conflict detection.
			conflicts, err = storage.ConflictingFilesForGist(g, prefs)
		}
		if err != nil {
			return errMsg{fmt.Errorf("conflict check failed: %w", err)}
		}
		return gistConflictCheckMsg{g: g, prefs: prefs, conflicts: conflicts}
	}
}

func (m GistModel) restoreGistCmd(g *gist.Gist, prefs storage.SyncPrefs, force bool) tea.Cmd {
	mgr := m.gistSyncMgr
	return func() tea.Msg {
		if mgr != nil {
			if err := mgr.PullFromGist(g, prefs, force); err != nil {
				return errMsg{fmt.Errorf("gist restore failed: %w", err)}
			}
			return successMsg{"✓ Data restored from Gist."}
		}
		// No token/sync manager — use standalone restore.
		if err := storage.RestoreFromGistDirect(g, prefs, force); err != nil {
			return errMsg{fmt.Errorf("gist restore failed: %w", err)}
		}
		return successMsg{"✓ Data restored from Gist."}
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m GistModel) View() string {
	if m.quitting {
		return ""
	}
	h := m.height

	switch m.state {
	case gistStateMenu:
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Sync & Backup",
			Subtitle: "Select an Option",
			Content:  m.viewMenuWithSections() + "\n" + m.renderMessage(),
			Help:     "↑↓/jk: Navigate • Enter/1-9/a/t: Select • Esc: Back",
		}, h)

	case gistStateCreateVisibility:
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Create Favorites Gist",
			Subtitle: "Choose Visibility",
			Content:  m.visibilityMenu.View() + "\n" + m.renderMessage(),
			Help:     "↑↓/jk: Navigate • Enter: Select • Esc: Back",
		}, h)

	case gistStateCreateName:
		vis := "Secret"
		if m.createPublic {
			vis = "Public"
		}
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Create Favorites Gist",
			Subtitle: fmt.Sprintf("Enter Gist Name (%s)", vis),
			Content:  fmt.Sprintf("Name/description for your gist:\n\n%s", m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Create • Esc: Back",
		}, h)

	case gistStateCreate:
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Create Favorites Gist",
			Subtitle: "Uploading…",
			Content:  "Creating gist, please wait…\n\n" + m.renderMessage(),
			Help:     "Please wait…",
		}, h)

	case gistStateRestoreGistURL:
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Restore All Data from Gist",
			Subtitle: "Paste a gist URL or ID",
			Content: fmt.Sprintf(
				"Enter a gist URL or ID:\n\n"+
					"  • https://gist.github.com/username/gist_id\n"+
					"  • Raw gist ID (e.g., abc123def456…)\n\n"+
					"Public gists work without a token.\n"+
					"Private gists require a token to be configured.\n\n"+
					"%s", m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help: "Enter: Fetch • Esc: Back",
		}, h)

	case gistStateImportURL:
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Import Favorites from URL",
			Subtitle: "Paste a public gist URL or ID",
			Content: fmt.Sprintf(
				"Enter a gist URL or ID:\n\n"+
					"  • https://gist.github.com/username/gist_id\n"+
					"  • Raw gist ID (e.g., abc123def456…)\n\n"+
					"%s", m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help: "Enter: Import • Esc: Back",
		}, h)

	case gistStateList, gistStateUpdate, gistStateDelete, gistStateRecover:
		action := "My Favorites Gists"
		switch m.state {
		case gistStateUpdate:
			action = "Update Favorites Gist"
		case gistStateDelete:
			action = "Delete Favorites Gist"
		case gistStateRecover:
			action = "Recover Favorites Gist"
		}
		content := m.gistList.View()
		if len(m.gists) == 0 {
			content = "No gists available.\n\nCreate a favorites gist first."
		}
		return RenderPageWithBottomHelp(PageLayout{
			Title:    action,
			Subtitle: "Select a Gist",
			Content:  content + "\n" + m.renderMessage(),
			Help:     "↑↓/jk: Navigate • Enter: Select • Esc: Back",
		}, h)

	case gistStateTokenMenu:
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Token Management",
			Subtitle: "Manage your GitHub Token",
			Content:  m.tokenMenuList.View() + "\n" + m.renderMessage(),
			Help:     "↑↓/jk: Navigate • Enter: Select • 1-4: Quick select • Esc: Back",
		}, h)

	case gistStateTokenSetup:
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Setup Token",
			Subtitle: "Paste your GitHub Token",
			Content:  fmt.Sprintf("Token will be hidden for security.\n\n%s", m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Save • Esc: Cancel",
		}, h)

	case gistStateTokenView:
		masked := gist.GetMaskedToken(m.token)
		sourceInfo := ""
		if m.token == "" {
			masked = "No token configured"
		} else {
			if source, err := gist.GetTokenSource(); err == nil {
				switch source {
				case gist.SourceKeychain:
					sourceInfo = "\nStorage: OS Keychain (secure)"
				case gist.SourceEnvironment:
					sourceInfo = "\nStorage: Environment Variable (TERA_GITHUB_TOKEN)"
				case gist.SourceFile:
					sourceInfo = "\nStorage: File (legacy)"
				}
			}
		}
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Current Token",
			Subtitle: "View Token Status",
			Content:  fmt.Sprintf("Token: %s%s", masked, sourceInfo) + "\n\n" + m.renderMessage(),
			Help:     "Enter/Esc: Back",
		}, h)

	case gistStateUpdateInput:
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Update Favorites Gist",
			Subtitle: "Enter new description",
			Content:  fmt.Sprintf("Current: %s\n\nNew Description:\n%s", m.selectedGist.Description, m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Update • Esc: Cancel",
		}, h)

	case gistStateDeleteConfirm:
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Delete Favorites Gist",
			Subtitle: "Confirm Deletion",
			Content:  fmt.Sprintf("Delete gist?\n%s\n\nType 'yes' to confirm:\n%s", m.selectedGist.Description, m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Confirm • Esc: Cancel",
		}, h)

	case gistStateTokenDelete:
		masked := gist.GetMaskedToken(m.token)
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Delete Token",
			Subtitle: "⚠️  WARNING",
			Content:  fmt.Sprintf("This will delete your stored GitHub token!\n\nToken: %s\n\nType 'yes' to confirm:\n%s", masked, m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Confirm • Esc: Cancel",
		}, h)

	// ── Phase 3 views ──────────────────────────────────────────────────────────

	case gistStateExportChecklist, gistStateRestoreZipChecklist,
		gistStateSyncGistChecklist, gistStateRestoreGistChecklist:
		return RenderPageWithBottomHelp(PageLayout{
			Title:   "Sync & Backup",
			Content: m.checklist.View() + m.renderMessage(),
			Help:    m.checklist.HelpText(),
		}, h)

	case gistStateExportPath:
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Export Backup",
			Subtitle: "Save location",
			Content: fmt.Sprintf(
				"Enter path for the zip file:\n\n%s\n\n%s",
				m.textInput.View(),
				dimStyle().Render("Tip: enter a directory to use the default filename"),
			) + "\n\n" + m.renderMessage(),
			Help: "Enter: Export • Esc: Back",
		}, h)

	case gistStateRestoreZipPath:
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Restore from Backup",
			Subtitle: "Zip file location",
			Content:  fmt.Sprintf("Enter path to the backup zip:\n\n%s", m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Open • Esc: Cancel",
		}, h)

	case gistStateOverwriteWarn:
		paths := strings.Join(m.overwritePaths, "\n  ")
		return RenderPageWithBottomHelp(PageLayout{
			Title:    "Overwrite Warning",
			Subtitle: "⚠️  The following files already exist:",
			Content:  fmt.Sprintf("  %s\n\nPress Enter to overwrite, Esc to cancel.", paths) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Overwrite • Esc: Cancel",
		}, h)

	case gistStateSyncProgress:
		return RenderPageWithBottomHelp(PageLayout{
			Title:   "Sync & Backup",
			Content: "Working…\n\n" + m.renderMessage(),
		}, h)
	}

	return ""
}

// viewMenuWithSections renders the menu with visual section dividers.
func (m GistModel) viewMenuWithSections() string {
	var b strings.Builder

	items := m.menuList.Items()
	cursor := m.menuList.Index()

	sections := []struct {
		beforeIdx int
		label     string
	}{
		{0, "— Favorites Gist —"},
		{6, "— Full Backup —"},
		{10, "— Account —"},
	}

	sIdx := 0
	for i, item := range items {
		for sIdx < len(sections) && sections[sIdx].beforeIdx == i {
			b.WriteString(dimStyle().Render("  "+sections[sIdx].label) + "\n")
			sIdx++
		}
		mi, ok := item.(components.MenuItem)
		if !ok {
			continue
		}
		label := mi.Title()
		if s := mi.Shortcut(); s != "" {
			label = fmt.Sprintf("%s. %s", s, label)
		}
		if i == cursor {
			b.WriteString(selectedItemStyle().Render("> " + label))
		} else {
			b.WriteString(normalItemStyle().Render("  " + label))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m GistModel) renderMessage() string {
	if m.message == "" {
		return ""
	}
	if m.messageIsError {
		return errorStyle().Render(m.message)
	}
	return successStyle().Render(m.message)
}

// ── Commands (pre-existing favorites Gist flows) ───────────────────────────────

func (m *GistModel) initCreateGist() (tea.Model, tea.Cmd) {
	if m.token == "" {
		m.message = "No token configured!"
		m.messageIsError = true
		m.state = gistStateMenu
		return *m, nil
	}
	m.message = "Creating gist…"
	return *m, m.createGistCmd
}

func (m *GistModel) createGistCmd() tea.Msg {
	files := make(map[string]*string)
	entries, err := os.ReadDir(m.favoritePath)
	if err != nil {
		return errMsg{err}
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			if content, err := os.ReadFile(filepath.Join(m.favoritePath, entry.Name())); err == nil {
				s := string(content)
				files[entry.Name()] = &s
			}
		}
	}
	if len(files) == 0 {
		return errMsg{fmt.Errorf("no favorite lists found in %s", m.favoritePath)}
	}
	description := m.gistDescription
	if description == "" {
		description = fmt.Sprintf("TERA Radio Favorites - %s", time.Now().Format("2006-01-02 15:04:05"))
	}
	newGist, err := m.gistClient.CreateGist(description, files, m.createPublic)
	if err != nil {
		return errMsg{err}
	}
	meta := &gist.GistMetadata{
		ID: newGist.ID, URL: newGist.URL, Description: newGist.Description,
		CreatedAt: newGist.CreatedAt, UpdatedAt: newGist.UpdatedAt,
	}
	if err := gist.SaveMetadata(meta); err != nil {
		return errMsg{err}
	}
	vis := "secret"
	if m.createPublic {
		vis = "public"
	}
	return successMsg{fmt.Sprintf("Gist created (%s)! %s", vis, newGist.URL)}
}

func (m *GistModel) initGistList() (tea.Model, tea.Cmd) {
	if m.token == "" || m.gistClient == nil {
		m.message = "No GitHub token configured. Set up a token first."
		m.messageIsError = true
		m.state = gistStateMenu
		return *m, nil
	}
	return *m, func() tea.Msg {
		gists, err := gist.GetAllGists()
		if err != nil {
			return errMsg{err}
		}
		if len(gists) == 0 {
			return errMsg{fmt.Errorf("no gists found — create a favorites gist first")}
		}
		return gistsMsg(gists)
	}
}

func (m *GistModel) initTokenMenu() (tea.Model, tea.Cmd) { return *m, nil }

func (m *GistModel) saveTokenCmd(token string) tea.Cmd {
	return func() tea.Msg {
		client := gist.NewClient(token)
		user, err := client.ValidateToken()
		if err != nil {
			return errMsg{fmt.Errorf("invalid token: %v", err)}
		}
		if err := gist.SaveToken(token); err != nil {
			return errMsg{err}
		}
		return tokenSavedMsg{token, user}
	}
}

func (m *GistModel) recoverGistCmd(gistID string) tea.Cmd {
	return func() tea.Msg {
		g, err := m.gistClient.GetGist(gistID)
		if err != nil {
			return errMsg{err}
		}
		backupDir := filepath.Join(m.favoritePath, ".backup")
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return errMsg{fmt.Errorf("failed to create backup directory: %w", err)}
		}
		timestamp := time.Now().Format("20060102-150405")
		var backupFailures []string
		for filename := range g.Files {
			cleanName := filepath.Base(filename)
			if cleanName != filename || cleanName == "." || cleanName == ".." {
				continue
			}
			existingPath := filepath.Join(m.favoritePath, cleanName)
			if _, err := os.Stat(existingPath); err == nil {
				backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.%s.bak", cleanName, timestamp))
				if data, err := os.ReadFile(existingPath); err == nil {
					if err := os.WriteFile(backupPath, data, 0644); err != nil {
						backupFailures = append(backupFailures, cleanName)
					}
				} else {
					backupFailures = append(backupFailures, cleanName)
				}
			}
		}
		for filename, file := range g.Files {
			cleanName := filepath.Base(filename)
			if cleanName != filename || cleanName == "." || cleanName == ".." {
				continue
			}
			if err := os.WriteFile(filepath.Join(m.favoritePath, cleanName), []byte(file.Content), 0644); err != nil {
				return errMsg{err}
			}
		}
		if len(backupFailures) > 0 {
			return successMsg{fmt.Sprintf("Favorites restored! (Warning: %d backup(s) failed)", len(backupFailures))}
		}
		return successMsg{"Favorites restored successfully! (backups saved in .backup folder)"}
	}
}

func (m *GistModel) importGistCmd(gistID string) tea.Cmd {
	return func() tea.Msg {
		var g *gist.Gist
		var err error
		if m.gistClient != nil {
			g, err = m.gistClient.GetGist(gistID)
		}
		if g == nil || err != nil {
			g, err = gist.GetGistPublic(gistID)
			if err != nil {
				return errMsg{fmt.Errorf("failed to fetch gist: %w", err)}
			}
		}
		backupDir := filepath.Join(m.favoritePath, ".backup")
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return errMsg{fmt.Errorf("failed to create backup directory: %w", err)}
		}
		timestamp := time.Now().Format("20060102-150405")
		for filename := range g.Files {
			cleanName := filepath.Base(filename)
			if cleanName != filename || cleanName == "." || cleanName == ".." {
				continue
			}
			existingPath := filepath.Join(m.favoritePath, cleanName)
			if _, err := os.Stat(existingPath); err == nil {
				backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.%s.bak", cleanName, timestamp))
				if data, err := os.ReadFile(existingPath); err == nil {
					_ = os.WriteFile(backupPath, data, 0644)
				}
			}
		}
		importedCount := 0
		for filename, file := range g.Files {
			cleanName := filepath.Base(filename)
			if cleanName != filename || cleanName == "." || cleanName == ".." {
				continue
			}
			if err := os.WriteFile(filepath.Join(m.favoritePath, cleanName), []byte(file.Content), 0644); err != nil {
				return errMsg{err}
			}
			importedCount++
		}
		return successMsg{fmt.Sprintf("Successfully imported %d file(s) from gist!", importedCount)}
	}
}

func (m *GistModel) updateGistCmd(id, description string) tea.Cmd {
	return func() tea.Msg {
		if err := m.gistClient.UpdateGist(id, description); err != nil {
			return errMsg{err}
		}
		if err := gist.UpdateMetadata(id, description); err != nil {
			return errMsg{fmt.Errorf("gist updated remotely but local cache failed: %w", err)}
		}
		return successMsg{"Gist updated!"}
	}
}

func (m *GistModel) validateTokenCmd() tea.Cmd {
	return func() tea.Msg {
		client := gist.NewClient(m.token)
		user, err := client.ValidateToken()
		if err != nil {
			return errMsg{fmt.Errorf("token validation failed: %v", err)}
		}
		return successMsg{fmt.Sprintf("✓ Token is VALID! GitHub user: %s", user)}
	}
}

func (m *GistModel) deleteGistCmd(id string) tea.Cmd {
	return func() tea.Msg {
		if m.gistClient == nil {
			return errMsg{fmt.Errorf("no gist client configured")}
		}
		if err := m.gistClient.DeleteGist(id); err != nil {
			return errMsg{err}
		}
		if err := gist.DeleteMetadata(id); err != nil {
			return errMsg{fmt.Errorf("gist deleted remotely but local cache failed: %w", err)}
		}
		return successMsg{"Gist deleted!"}
	}
}

func (m *GistModel) deleteTokenCmd() tea.Cmd {
	return func() tea.Msg {
		if _, err := gist.DeleteToken(); err != nil {
			return errMsg{fmt.Errorf("failed to delete token: %v", err)}
		}
		return tokenDeletedMsg{}
	}
}

// openBrowser opens url in the default system browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
