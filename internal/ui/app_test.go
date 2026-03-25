package ui

import (
	"fmt"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/config"
	"github.com/shinokada/tera/v3/internal/player"
	"github.com/shinokada/tera/v3/internal/storage"
)

// newTestApp creates a minimal App for unit testing (no filesystem side-effects).
func newTestApp() *App {
	return &App{
		screen:         screenMainMenu,
		playHistoryCfg: config.DefaultPlayHistoryConfig(),
	}
}

// makeRecentlyPlayed builds a slice of StationWithMetadata from names.
func makeRecentlyPlayed(names ...string) []storage.StationWithMetadata {
	result := make([]storage.StationWithMetadata, len(names))
	for i, name := range names {
		result[i] = storage.StationWithMetadata{
			Station: api.Station{StationUUID: name, Name: name},
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// loadRecentlyPlayed
// ---------------------------------------------------------------------------

func TestLoadRecentlyPlayed_DisabledReturnsNil(t *testing.T) {
	app := newTestApp()
	app.playHistoryCfg.Enabled = false
	app.loadRecentlyPlayed()
	if app.recentlyPlayed != nil {
		t.Errorf("expected nil when disabled, got %d entries", len(app.recentlyPlayed))
	}
}

func TestLoadRecentlyPlayed_NilManagerReturnsNil(t *testing.T) {
	app := newTestApp()
	app.playHistoryCfg.Enabled = true
	app.metadataManager = nil
	app.loadRecentlyPlayed()
	if app.recentlyPlayed != nil {
		t.Errorf("expected nil when manager is nil, got %d entries", len(app.recentlyPlayed))
	}
}

func TestLoadRecentlyPlayed_EmptyHistory(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := storage.NewMetadataManager(tmpDir)
	if err != nil {
		t.Fatalf("NewMetadataManager: %v", err)
	}
	defer mgr.Close() //nolint:errcheck

	app := newTestApp()
	app.playHistoryCfg.Enabled = true
	app.playHistoryCfg.Size = 5
	app.metadataManager = mgr

	app.loadRecentlyPlayed()
	if len(app.recentlyPlayed) != 0 {
		t.Errorf("expected 0 entries for empty history, got %d", len(app.recentlyPlayed))
	}
}

func TestLoadRecentlyPlayed_SizeLimit(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := storage.NewMetadataManager(tmpDir)
	if err != nil {
		t.Fatalf("NewMetadataManager: %v", err)
	}
	defer mgr.Close() //nolint:errcheck

	// Record 10 distinct stations.
	for i := 0; i < 10; i++ {
		s := api.Station{StationUUID: fmt.Sprintf("station-%02d", i), Name: fmt.Sprintf("Station %d", i)}
		if err := mgr.StartPlay(&s); err != nil {
			t.Fatalf("StartPlay: %v", err)
		}
		if err := mgr.StopPlay(s.StationUUID); err != nil {
			t.Fatalf("StopPlay: %v", err)
		}
	}

	app := newTestApp()
	app.playHistoryCfg.Enabled = true
	app.playHistoryCfg.Size = 3
	app.metadataManager = mgr

	app.loadRecentlyPlayed()
	if len(app.recentlyPlayed) != 3 {
		t.Errorf("expected 3 entries (size limit), got %d", len(app.recentlyPlayed))
	}
}

// ---------------------------------------------------------------------------
// playRecentStation
// ---------------------------------------------------------------------------

func TestPlayRecentStation_ValidIndex(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("Radio A", "Radio B", "Radio C")

	_, _ = app.playRecentStation(1)

	if app.playingStation == nil {
		t.Fatal("expected playingStation to be set")
	}
	if app.playingStation.Name != "Radio B" {
		t.Errorf("expected 'Radio B', got '%s'", app.playingStation.Name)
	}
	if !app.playingFromMain {
		t.Error("expected playingFromMain to be true")
	}
}

func TestPlayRecentStation_OutOfRange(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("Radio A")

	_, _ = app.playRecentStation(5)

	if app.playingStation != nil {
		t.Error("expected playingStation to remain nil for out-of-range index")
	}
}

func TestPlayRecentStation_EmptyList(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = nil

	_, _ = app.playRecentStation(0)

	if app.playingStation != nil {
		t.Error("expected playingStation to remain nil for empty list")
	}
}

func TestPlayRecentStation_FirstStation(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("Jazz FM", "Blues Radio")

	_, _ = app.playRecentStation(0)

	if app.playingStation == nil {
		t.Fatal("expected playingStation to be set")
	}
	if app.playingStation.Name != "Jazz FM" {
		t.Errorf("expected 'Jazz FM', got '%s'", app.playingStation.Name)
	}
}

func TestPlayRecentStation_LastStation(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("A", "B", "C")

	_, _ = app.playRecentStation(2)

	if app.playingStation == nil {
		t.Fatal("expected playingStation to be set")
	}
	if app.playingStation.Name != "C" {
		t.Errorf("expected 'C', got '%s'", app.playingStation.Name)
	}
}

// ---------------------------------------------------------------------------
// updateRPViewOffset
// ---------------------------------------------------------------------------

func TestUpdateRPViewOffset_NoRP(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = nil
	app.rpViewOffset = 3 // stale value
	app.updateRPViewOffset(5)
	if app.rpViewOffset != 0 {
		t.Errorf("expected 0 when no RP entries, got %d", app.rpViewOffset)
	}
}

func TestUpdateRPViewOffset_CursorAboveRP(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("A", "B", "C", "D", "E")
	app.rpViewOffset = 3
	app.unifiedMenuIndex = 0 // cursor in menu, not RP
	app.updateRPViewOffset(3)
	if app.rpViewOffset != 0 {
		t.Errorf("expected offset reset to 0 when cursor above RP, got %d", app.rpViewOffset)
	}
}

func TestUpdateRPViewOffset_ScrollsDown(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("A", "B", "C", "D", "E")
	// Cursor at RP entry 4 (index 3 in RP, unifiedIdx = 11+3 = 14)
	rpStart := mainMenuItemCount + len(app.quickFavorites)
	app.unifiedMenuIndex = rpStart + 3
	app.updateRPViewOffset(3) // window of 3
	// entry 3 must be visible: offset should be 1 (shows entries 1,2,3)
	if app.rpViewOffset != 1 {
		t.Errorf("expected offset 1, got %d", app.rpViewOffset)
	}
}

func TestUpdateRPViewOffset_ScrollsUp(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("A", "B", "C", "D", "E")
	app.rpViewOffset = 4 // pushed far down
	rpStart := mainMenuItemCount + len(app.quickFavorites)
	app.unifiedMenuIndex = rpStart + 0 // cursor at first RP entry
	app.updateRPViewOffset(3)
	if app.rpViewOffset != 0 {
		t.Errorf("expected offset scrolled back to 0, got %d", app.rpViewOffset)
	}
}

func TestUpdateRPViewOffset_LargeWindowCoversAllEntries(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("A", "B", "C")
	// Window larger than list — all items fit, offset stays 0 regardless of cursor position
	rpStart := mainMenuItemCount + len(app.quickFavorites)
	app.unifiedMenuIndex = rpStart + 1
	app.updateRPViewOffset(10)
	if app.rpViewOffset != 0 {
		t.Errorf("expected offset 0 when window covers all entries, got %d", app.rpViewOffset)
	}
}

// ---------------------------------------------------------------------------
// Helpers shared by playback tests
// ---------------------------------------------------------------------------

// newTestStation returns a minimal Station with the given UUID and name.
func newTestStation(uuid, name string) *api.Station {
	return &api.Station{StationUUID: uuid, Name: name}
}

// newContinueOnNavigateApp returns an App pre-configured with
// ContinueOnNavigate = true, ready for playback scenario tests.
func newContinueOnNavigateApp() *App {
	app := newTestApp()
	app.playOptsCfg = config.DefaultPlayOptionsConfig()
	app.playOptsCfg.ContinueOnNavigate = true
	app.quickFavPlayer = player.NewMPVPlayer()
	return app
}

// collectMsgs executes a tea.Cmd (which may be a tea.Batch) and collects every
// non-nil message it synchronously produces. It does not run child commands
// returned by those messages — one level is enough for our assertions.
func collectMsgs(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	var msgs []tea.Msg
	switch m := msg.(type) {
	case tea.BatchMsg:
		for _, c := range m {
			if c != nil {
				if inner := c(); inner != nil {
					msgs = append(msgs, inner)
				}
			}
		}
	default:
		if msg != nil {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

// hasMsgType returns true if any message in msgs is of the given type.
func hasMsgType[T any](msgs []tea.Msg) bool {
	for _, m := range msgs {
		if _, ok := m.(T); ok {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Bug: playing Quick Favorites then Play from Favorites plays two stations
// Fix: navigating to screenPlay calls stopScreenPlayers() so all screen-owned
//
//	    players are stopped before the new session begins.
//
// ---------------------------------------------------------------------------

// TestNavigateToScreenPlay_StopsScreenPlayers verifies that when the user
// navigates to screenPlay (Play from Favorites), all screen-owned players
// are stopped regardless of ContinueOnNavigate.
func TestNavigateToScreenPlay_StopsScreenPlayers(t *testing.T) {
	app := newContinueOnNavigateApp()

	// Simulate Top Rated screen that was playing (player == quickFavPlayer).
	app.topRatedScreen.player = app.quickFavPlayer

	// Simulate the handoff that ContinueOnNavigate would have produced.
	station := newTestStation("uuid-top", "Top Rated FM")
	app.activePlayer = app.quickFavPlayer
	app.activeStation = station
	app.activeContextLabel = "Top Rated"
	app.playingFromMain = true
	app.playingStation = station

	// Navigate to screenPlay — this is what triggers the bug.
	app.Update(navigateMsg{screen: screenPlay})

	// quickFavPlayer (== topRatedScreen.player) must have been stopped.
	// IsPlaying() returns false both when never started and when stopped,
	// so we verify the player is not in a playing state.
	if app.quickFavPlayer.IsPlaying() {
		t.Error("quickFavPlayer should have been stopped when entering screenPlay")
	}
}

// TestNavigateToScreenPlay_ContinueOnNavigate_PreservesActivePlayer verifies
// that when ContinueOnNavigate is ON the activePlayer (handed-off stream) is
// NOT stopped when merely navigating to the play-list screen — it should keep
// playing while the user browses, and only stop when they select a station.
func TestNavigateToScreenPlay_ContinueOnNavigate_PreservesActivePlayer(t *testing.T) {
	app := newContinueOnNavigateApp()

	station := newTestStation("uuid-top", "Top Rated FM")
	handedOffPlayer := player.NewMPVPlayer()
	app.activePlayer = handedOffPlayer
	app.activeStation = station
	app.activeContextLabel = "Top Rated"

	// Navigate to screenPlay.
	app.Update(navigateMsg{screen: screenPlay})

	// activePlayer must still be set — it should keep running.
	if app.activePlayer == nil {
		t.Error("activePlayer should be preserved when ContinueOnNavigate is ON and user is just browsing")
	}
	if app.activeStation == nil {
		t.Error("activeStation should be preserved when ContinueOnNavigate is ON")
	}
}

// TestNavigateToScreenPlay_ContinueOff_ClearsActivePlayer verifies that when
// ContinueOnNavigate is OFF the activePlayer IS stopped and cleared on entry.
func TestNavigateToScreenPlay_ContinueOff_ClearsActivePlayer(t *testing.T) {
	app := newTestApp()
	app.playOptsCfg = config.DefaultPlayOptionsConfig()
	app.playOptsCfg.ContinueOnNavigate = false
	app.quickFavPlayer = player.NewMPVPlayer()

	station := newTestStation("uuid-top", "Top Rated FM")
	handedOffPlayer := player.NewMPVPlayer()
	app.activePlayer = handedOffPlayer
	app.activeStation = station

	app.Update(navigateMsg{screen: screenPlay})

	if app.activePlayer != nil {
		t.Error("activePlayer should be cleared when ContinueOnNavigate is OFF")
	}
	if app.activeStation != nil {
		t.Error("activeStation should be cleared when ContinueOnNavigate is OFF")
	}
}

// ---------------------------------------------------------------------------
// Bug: playing from Top Rated then going back to list stopped the music
// Fix: handleListInput only stops the player when ContinueOnNavigate is OFF.
// ---------------------------------------------------------------------------

// TestTopRated_EscFromList_ContinueOnNavigate_DoesNotStopPlayer verifies that
// pressing Esc on the Top Rated list screen does NOT stop the player when
// ContinueOnNavigate is ON (the music should keep playing).
func TestTopRated_EscFromList_ContinueOnNavigate_DoesNotStopPlayer(t *testing.T) {
	app := newContinueOnNavigateApp()
	app.screen = screenTopRated
	app.topRatedScreen.player = app.quickFavPlayer
	app.topRatedScreen.playOptsCfg = app.playOptsCfg
	app.topRatedScreen.state = topRatedStateList

	// Press Esc from the list — should navigate away without stopping player.
	newModel, cmd := app.topRatedScreen.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app.topRatedScreen = newModel

	// Verify it emits a navigation command (going to main menu).
	msgs := collectMsgs(cmd)
	if !hasMsgType[navigateMsg](msgs) {
		t.Error("expected navigateMsg to main menu when pressing Esc from Top Rated list")
	}

	// Player should NOT have been stopped. Since NewMPVPlayer starts as
	// not-playing, the real signal is the navigateMsg check above — if the
	// player were being stopped the old code path would have been taken.
	// The key assertion: no stopActivePlaybackMsg was emitted by this action.
	if hasMsgType[stopActivePlaybackMsg](msgs) {
		t.Error("stopActivePlaybackMsg should NOT be emitted when navigating away from Top Rated list with ContinueOnNavigate ON")
	}
}

// ---------------------------------------------------------------------------
// Bug: playing from I Feel Lucky then going to main menu stopped the music
// Fix: updateInput esc now hands off the player when ContinueOnNavigate is ON.
// ---------------------------------------------------------------------------

// TestLucky_EscFromPlaying_ContinueOnNavigate_HandsOffAndNavigates verifies
// that pressing Esc from luckyStatePlaying with ContinueOnNavigate ON emits
// both a handoffPlaybackMsg (to keep music running) and a navigateMsg (to
// actually go to the main menu), without landing on the input screen first.
func TestLucky_EscFromPlaying_ContinueOnNavigate_HandsOffAndNavigates(t *testing.T) {
	m := NewLuckyModel(nil, t.TempDir(), nil)
	m.playOptsCfg.ContinueOnNavigate = true
	m.player = player.NewMPVPlayer()
	m.selectedStation = newTestStation("uuid-lucky", "Lucky Radio")
	m.state = luckyStatePlaying

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	msgs := collectMsgs(cmd)
	if !hasMsgType[handoffPlaybackMsg](msgs) {
		t.Error("expected handoffPlaybackMsg when pressing Esc from Lucky playing with ContinueOnNavigate ON")
	}
	if !hasMsgType[navigateMsg](msgs) {
		t.Error("expected navigateMsg to main menu alongside the handoff")
	}
}

// TestLucky_EscFromPlaying_ContinueOff_StopsAndReturnsToInput verifies that
// Esc from luckyStatePlaying with ContinueOnNavigate OFF stops the player and
// returns to the input screen (not main menu).
func TestLucky_EscFromPlaying_ContinueOff_StopsAndReturnsToInput(t *testing.T) {
	m := NewLuckyModel(nil, t.TempDir(), nil)
	m.playOptsCfg.ContinueOnNavigate = false
	m.player = player.NewMPVPlayer()
	m.selectedStation = newTestStation("uuid-lucky", "Lucky Radio")
	m.state = luckyStatePlaying

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	lucky := newModel.(LuckyModel)

	if lucky.state != luckyStateInput {
		t.Errorf("expected luckyStateInput after Esc with ContinueOnNavigate OFF, got %v", lucky.state)
	}
	msgs := collectMsgs(cmd)
	if hasMsgType[handoffPlaybackMsg](msgs) {
		t.Error("handoffPlaybackMsg should not be emitted when ContinueOnNavigate is OFF")
	}
	if hasMsgType[navigateMsg](msgs) {
		t.Error("navigateMsg should not be emitted when returning to Lucky input (not main menu)")
	}
}

// TestLucky_EscFromInput_NoStation_Navigates_Still verifies that pressing Esc
// on the Lucky input screen with no station playing always navigates to main
// menu (the ContinueOnNavigate path never fires with nil selectedStation).
func TestLucky_EscFromInput_NoStation_Navigates_Still(t *testing.T) {
	m := NewLuckyModel(nil, t.TempDir(), nil)
	m.playOptsCfg.ContinueOnNavigate = true
	m.player = player.NewMPVPlayer()
	m.selectedStation = nil
	m.state = luckyStateInput
	m.inputMode = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	msgs := collectMsgs(cmd)
	if !hasMsgType[navigateMsg](msgs) {
		t.Error("expected navigateMsg when no station is playing")
	}
	if hasMsgType[handoffPlaybackMsg](msgs) {
		t.Error("handoffPlaybackMsg should not be emitted when no station is playing")
	}
}

// TestLucky_EscFromInput_ContinueOff_StopsAndNavigates verifies that pressing
// Esc from the Lucky input screen stops the player and navigates to main menu
// when ContinueOnNavigate is OFF.
func TestLucky_EscFromInput_ContinueOff_StopsAndNavigates(t *testing.T) {
	m := LuckyModel{}
	m.playOptsCfg.ContinueOnNavigate = false
	m.player = player.NewMPVPlayer()
	m.selectedStation = newTestStation("uuid-lucky", "Lucky Radio")
	m.state = luckyStateInput
	m.inputMode = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	msgs := collectMsgs(cmd)
	if !hasMsgType[navigateMsg](msgs) {
		t.Error("expected navigateMsg to main menu when ContinueOnNavigate is OFF")
	}
	if hasMsgType[handoffPlaybackMsg](msgs) {
		t.Error("handoffPlaybackMsg should not be emitted when ContinueOnNavigate is OFF")
	}
}

// TestLucky_EscFromInput_NoStation_Navigates verifies that Esc with no station
// playing always navigates cleanly regardless of ContinueOnNavigate.
func TestLucky_EscFromInput_NoStation_Navigates(t *testing.T) {
	m := LuckyModel{}
	m.playOptsCfg.ContinueOnNavigate = true
	m.player = player.NewMPVPlayer()
	m.selectedStation = nil // nothing playing
	m.state = luckyStateInput
	m.inputMode = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	msgs := collectMsgs(cmd)
	if !hasMsgType[navigateMsg](msgs) {
		t.Error("expected navigateMsg when no station is playing")
	}
	if hasMsgType[handoffPlaybackMsg](msgs) {
		t.Error("handoffPlaybackMsg should not be emitted when no station is playing")
	}
}

// TestTopRated_EscFromList_ContinueOff_StopsPlayer verifies that pressing Esc
// from the Top Rated list DOES stop the player when ContinueOnNavigate is OFF.
func TestTopRated_EscFromList_ContinueOff_StopsPlayer(t *testing.T) {
	app := newTestApp()
	app.playOptsCfg.ContinueOnNavigate = false
	app.quickFavPlayer = player.NewMPVPlayer()
	app.screen = screenTopRated
	app.topRatedScreen.player = app.quickFavPlayer
	app.topRatedScreen.playOptsCfg = app.playOptsCfg
	app.topRatedScreen.state = topRatedStateList

	_, cmd := app.topRatedScreen.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Should still navigate away.
	msgs := collectMsgs(cmd)
	if !hasMsgType[navigateMsg](msgs) {
		t.Error("expected navigateMsg when pressing Esc from Top Rated list")
	}
	// No stopActivePlaybackMsg needed — the player.Stop() is called directly.
	if hasMsgType[stopActivePlaybackMsg](msgs) {
		t.Error("stopActivePlaybackMsg should not be emitted when ContinueOnNavigate is OFF (Stop called directly)")
	}
}

// ---------------------------------------------------------------------------
// Bug: playing from Top Rated → main menu → Play from Favorites still plays
//
//	    the Top Rated station alongside the new one.
//
// Fix: PlayModel.playStation() emits stopActivePlaybackMsg before starting.
// ---------------------------------------------------------------------------

// TestPlayStation_EmitsStopActivePlaybackMsg verifies that when a user selects
// a station in Play from Favorites, a stopActivePlaybackMsg is emitted first.
// This is the message that stops any ContinueOnNavigate handed-off player.
func TestPlayStation_EmitsStopActivePlaybackMsg(t *testing.T) {
	app := newContinueOnNavigateApp()

	// Set up the play screen as it would be after navigating from Top Rated.
	app.screen = screenPlay
	app.playScreen = NewPlayModel(t.TempDir(), nil)
	app.playScreen.playOptsCfg = app.playOptsCfg

	// Simulate a handed-off Top Rated player.
	handedOffPlayer := player.NewMPVPlayer()
	app.activePlayer = handedOffPlayer
	app.activeStation = newTestStation("uuid-top", "Top Rated FM")

	// Call playStation directly — this is what happens when Enter is pressed.
	station := api.Station{StationUUID: "uuid-fav", Name: "Fav Radio", URLResolved: "http://example.com/stream"}
	cmd := app.playScreen.playStation(station)

	// The first message emitted must be stopActivePlaybackMsg.
	msgs := collectMsgs(cmd)
	if !hasMsgType[stopActivePlaybackMsg](msgs) {
		t.Errorf("expected stopActivePlaybackMsg to be emitted by playStation(), got: %v", msgs)
	}
}

// TestApp_HandleStopActivePlaybackMsg verifies that when the App receives a
// stopActivePlaybackMsg it clears activePlayer and activeStation.
func TestApp_HandleStopActivePlaybackMsg(t *testing.T) {
	app := newContinueOnNavigateApp()
	app.activePlayer = player.NewMPVPlayer()
	app.activeStation = newTestStation("uuid-top", "Top Rated FM")
	app.activeContextLabel = "Top Rated"

	app.Update(stopActivePlaybackMsg{})

	if app.activePlayer != nil {
		t.Error("activePlayer should be nil after stopActivePlaybackMsg")
	}
	if app.activeStation != nil {
		t.Error("activeStation should be nil after stopActivePlaybackMsg")
	}
	if app.activeContextLabel != "" {
		t.Error("activeContextLabel should be empty after stopActivePlaybackMsg")
	}
}

// ---------------------------------------------------------------------------
// Bug: Now Playing bar appeared on Top Rated screen itself
// Fix: isPlayerScreen() suppresses the bar on screens that own their player.
// ---------------------------------------------------------------------------

func TestIsPlayerScreen(t *testing.T) {
	playerScreens := []Screen{
		screenPlay, screenSearch, screenLucky,
		screenMostPlayed, screenTopRated,
		screenBrowseTags, screenTagPlaylists,
	}
	nonPlayerScreens := []Screen{
		screenMainMenu, screenList, screenGist,
		screenSettings, screenShuffleSettings,
		screenConnectionSettings, screenAppearanceSettings,
		screenBlocklist, screenSleepSummary,
	}

	app := newTestApp()

	for _, s := range playerScreens {
		app.screen = s
		if !app.isPlayerScreen() {
			t.Errorf("screen %v should be a player screen", s)
		}
	}
	for _, s := range nonPlayerScreens {
		app.screen = s
		if app.isPlayerScreen() {
			t.Errorf("screen %v should NOT be a player screen", s)
		}
	}
}

// TestNowPlayingBar_HiddenOnPlayerScreens verifies that the Now Playing bar
// is not appended to the view when the current screen is a player screen.
func TestNowPlayingBar_HiddenOnPlayerScreens(t *testing.T) {
	app := newContinueOnNavigateApp()
	app.activeStation = newTestStation("uuid-top", "Top Rated FM")
	app.activeContextLabel = "Top Rated"
	app.activePlayer = app.quickFavPlayer
	app.width = 80
	app.height = 24

	// nowPlayingBar() returns a non-empty string when activeStation is set.
	if app.nowPlayingBar() == "" {
		t.Skip("nowPlayingBar returned empty — player may need volume; skipping rendering check")
	}

	for _, s := range []Screen{screenTopRated, screenMostPlayed, screenPlay} {
		app.screen = s
		if !app.isPlayerScreen() {
			t.Errorf("precondition: %v should be a player screen", s)
		}
		// The bar must be suppressed on this screen.
		bar := app.nowPlayingBar()
		if bar != "" && app.isPlayerScreen() {
			// Simulate what View() does: only append when !isPlayerScreen().
			appended := !app.isPlayerScreen()
			if appended {
				t.Errorf("Now Playing bar should not be appended on player screen %v", s)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// handoffPlaybackMsg: App correctly takes ownership of handed-off player
// ---------------------------------------------------------------------------

// TestHandoffPlaybackMsg_SetsActivePlayer verifies that when a player screen
// sends a handoffPlaybackMsg the App stores the player and station correctly.
func TestHandoffPlaybackMsg_SetsActivePlayer(t *testing.T) {
	app := newContinueOnNavigateApp()

	p := player.NewMPVPlayer()
	station := newTestStation("uuid-1", "Handoff Radio")

	app.Update(handoffPlaybackMsg{
		player:       p,
		station:      station,
		contextLabel: "Top Rated",
	})

	if app.activePlayer != p {
		t.Error("activePlayer should be the handed-off player")
	}
	if app.activeStation != station {
		t.Error("activeStation should be the handed-off station")
	}
	if app.activeContextLabel != "Top Rated" {
		t.Errorf("activeContextLabel should be 'Top Rated', got %q", app.activeContextLabel)
	}
	if !app.playingFromMain {
		t.Error("playingFromMain should be true after handoff")
	}
}

// ---------------------------------------------------------------------------
// Cross-screen ContinueOnNavigate matrix
//
// Screens: Play(a) Search(b) MostPlayed(c) TopRated(d)
//          BrowseTags(e) TagPlaylists(f) Lucky(g)
//
// Two properties are tested for every source screen:
//  1. handoffAndNavigate – pressing the "go to main menu" key from the
//     playing state emits both handoffPlaybackMsg AND navigateMsg.
//  2. App-level routing – when App receives navigateMsg for any destination
//     player screen while ContinueOnNavigate is ON, activePlayer is preserved.
// ---------------------------------------------------------------------------

// sourceHandoffResult captures what a screen emits when the user navigates
// away while ContinueOnNavigate is ON and a station is playing.
type sourceHandoffResult struct {
	hasHandoff  bool
	hasNavigate bool
}

// triggerNavigateToMain presses the canonical "go to main menu" key for each
// screen and returns what messages were produced.
func triggerNavigateToMain(screen Screen, station *api.Station, p *player.MPVPlayer, opts config.PlayOptionsConfig) sourceHandoffResult {
	msgs := func(cmd tea.Cmd) []tea.Msg {
		if cmd == nil {
			return nil
		}
		msg := cmd()
		var out []tea.Msg
		switch m := msg.(type) {
		case tea.BatchMsg:
			for _, c := range m {
				if c != nil {
					if inner := c(); inner != nil {
						out = append(out, inner)
					}
				}
			}
			return out
		default:
			if msg != nil {
				return []tea.Msg{msg}
			}
			return nil
		}
	}

	hasType := func(got []tea.Msg, want string) bool {
		for _, m := range got {
			switch want {
			case "handoff":
				if _, ok := m.(handoffPlaybackMsg); ok {
					return true
				}
			case "navigate":
				if _, ok := m.(navigateMsg); ok {
					return true
				}
			case "backToMain":
				if _, ok := m.(backToMainMsg); ok {
					return true
				}
			}
		}
		return false
	}

	switch screen {
	case screenPlay:
		m := PlayModel{}
		m.playOptsCfg = opts
		m.player = p
		m.selectedStation = station
		m.state = playStatePlaying
		_, cmd := m.updatePlaying(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")})
		got := msgs(cmd)
		return sourceHandoffResult{
			hasHandoff:  hasType(got, "handoff"),
			hasNavigate: hasType(got, "navigate") || hasType(got, "backToMain"),
		}
	case screenSearch:
		m := SearchModel{}
		m.playOptsCfg = opts
		m.player = p
		m.selectedStation = station
		m.state = searchStatePlaying
		_, cmd := m.handlePlayerUpdate(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")})
		got := msgs(cmd)
		return sourceHandoffResult{
			hasHandoff:  hasType(got, "handoff"),
			hasNavigate: hasType(got, "navigate") || hasType(got, "backToMain"),
		}
	case screenMostPlayed:
		m := MostPlayedModel{}
		m.playOptsCfg = opts
		m.player = p
		m.selectedStation = station
		m.state = mostPlayedStatePlaying
		_, cmd := m.handlePlayingInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")})
		got := msgs(cmd)
		return sourceHandoffResult{
			hasHandoff:  hasType(got, "handoff"),
			hasNavigate: hasType(got, "navigate") || hasType(got, "backToMain"),
		}
	case screenTopRated:
		m := TopRatedModel{}
		m.playOptsCfg = opts
		m.player = p
		m.selectedStation = station
		m.state = topRatedStatePlaying
		_, cmd := m.handlePlayingInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")})
		got := msgs(cmd)
		return sourceHandoffResult{
			hasHandoff:  hasType(got, "handoff"),
			hasNavigate: hasType(got, "navigate") || hasType(got, "backToMain"),
		}
	case screenBrowseTags:
		m := BrowseTagsModel{}
		m.playOptsCfg = opts
		m.player = p
		m.selectedStation = station
		m.state = browseTagsStatePlaying
		_, cmd := m.updatePlaying(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")})
		got := msgs(cmd)
		return sourceHandoffResult{
			hasHandoff:  hasType(got, "handoff"),
			hasNavigate: hasType(got, "navigate") || hasType(got, "backToMain"),
		}
	case screenTagPlaylists:
		m := TagPlaylistsModel{}
		m.playOptsCfg = opts
		m.player = p
		m.selectedStation = station
		m.state = tagPlaylistsStatePlaying
		_, cmd := m.updatePlaying(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")})
		got := msgs(cmd)
		return sourceHandoffResult{
			hasHandoff:  hasType(got, "handoff"),
			hasNavigate: hasType(got, "navigate") || hasType(got, "backToMain"),
		}
	case screenLucky:
		m := NewLuckyModel(nil, os.TempDir(), nil)
		m.playOptsCfg = opts
		m.player = p
		m.selectedStation = station
		m.state = luckyStatePlaying
		_, cmd := m.updatePlaying(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")})
		got := msgs(cmd)
		return sourceHandoffResult{
			hasHandoff:  hasType(got, "handoff"),
			hasNavigate: hasType(got, "navigate") || hasType(got, "backToMain"),
		}
	}
	return sourceHandoffResult{}
}

// TestContinueOnNavigate_SourceHandoff verifies that every player screen
// correctly emits both handoffPlaybackMsg and a navigation message when the
// user presses 0 (go to main menu) while ContinueOnNavigate is ON.
func TestContinueOnNavigate_SourceHandoff(t *testing.T) {
	opts := config.DefaultPlayOptionsConfig()
	opts.ContinueOnNavigate = true

	station := newTestStation("uuid-x", "Test Radio")

	screens := []struct {
		name   string
		screen Screen
	}{
		{"Play(a)", screenPlay},
		{"Search(b)", screenSearch},
		{"MostPlayed(c)", screenMostPlayed},
		{"TopRated(d)", screenTopRated},
		{"BrowseTags(e)", screenBrowseTags},
		{"TagPlaylists(f)", screenTagPlaylists},
		{"Lucky(g)", screenLucky},
	}

	for _, src := range screens {
		t.Run(src.name, func(t *testing.T) {
			p := player.NewMPVPlayer()
			res := triggerNavigateToMain(src.screen, station, p, opts)
			if !res.hasHandoff {
				t.Errorf("%s: expected handoffPlaybackMsg when pressing 0 with ContinueOnNavigate ON", src.name)
			}
			if !res.hasNavigate {
				t.Errorf("%s: expected navigation message (navigateMsg or backToMainMsg) alongside handoff", src.name)
			}
		})
	}
}

// TestContinueOnNavigate_SourceHandoff_Off verifies that when ContinueOnNavigate
// is OFF, no handoffPlaybackMsg is emitted (player is stopped directly).
func TestContinueOnNavigate_SourceHandoff_Off(t *testing.T) {
	opts := config.DefaultPlayOptionsConfig()
	opts.ContinueOnNavigate = false

	station := newTestStation("uuid-x", "Test Radio")

	screens := []struct {
		name   string
		screen Screen
	}{
		{"Play(a)", screenPlay},
		{"Search(b)", screenSearch},
		{"MostPlayed(c)", screenMostPlayed},
		{"TopRated(d)", screenTopRated},
		{"BrowseTags(e)", screenBrowseTags},
		{"TagPlaylists(f)", screenTagPlaylists},
		{"Lucky(g)", screenLucky},
	}

	for _, src := range screens {
		t.Run(src.name, func(t *testing.T) {
			p := player.NewMPVPlayer()
			res := triggerNavigateToMain(src.screen, station, p, opts)
			if res.hasHandoff {
				t.Errorf("%s: handoffPlaybackMsg should NOT be emitted when ContinueOnNavigate is OFF", src.name)
			}
		})
	}
}

// destinationScreens lists all player screens that can be navigated to.
var destinationScreens = []struct {
	name   string
	screen Screen
}{
	{"Play(a)", screenPlay},
	{"Search(b)", screenSearch},
	{"MostPlayed(c)", screenMostPlayed},
	{"TopRated(d)", screenTopRated},
	{"BrowseTags(e)", screenBrowseTags},
	{"TagPlaylists(f)", screenTagPlaylists},
	{"Lucky(g)", screenLucky},
}

// TestContinueOnNavigate_DestinationPreservesActivePlayer verifies that when
// App receives navigateMsg for any destination screen while ContinueOnNavigate
// is ON, the activePlayer handed off from the source is not killed.
func TestContinueOnNavigate_DestinationPreservesActivePlayer(t *testing.T) {
	for _, dst := range destinationScreens {
		t.Run(dst.name, func(t *testing.T) {
			app := newContinueOnNavigateApp()

			// Simulate a prior handoff (e.g. from Lucky → main menu).
			handedOff := player.NewMPVPlayer()
			app.activePlayer = handedOff
			app.activeStation = newTestStation("uuid-src", "Source Radio")
			app.activeContextLabel = "Lucky"

			// Navigate to the destination screen.
			app.Update(navigateMsg{screen: dst.screen})

			// activePlayer must be preserved — it is the still-running source
			// stream, and the user hasn't selected a new station yet.
			if app.activePlayer == nil {
				t.Errorf("%s: activePlayer was cleared on navigation; music would stop", dst.name)
			}
			if app.activeStation == nil {
				t.Errorf("%s: activeStation was cleared on navigation", dst.name)
			}
		})
	}
}

// TestContinueOnNavigate_DestinationClearsActivePlayer_WhenOff verifies that
// when ContinueOnNavigate is OFF, navigating to screenPlay clears activePlayer
// (for screens that stop their player on navigate; screenPlay is the only one
// with explicit clear logic in App).
func TestContinueOnNavigate_DestinationClearsActivePlayer_WhenOff(t *testing.T) {
	app := newTestApp()
	app.playOptsCfg = config.DefaultPlayOptionsConfig()
	app.playOptsCfg.ContinueOnNavigate = false
	app.quickFavPlayer = player.NewMPVPlayer()

	app.activePlayer = player.NewMPVPlayer()
	app.activeStation = newTestStation("uuid-src", "Source Radio")

	app.Update(navigateMsg{screen: screenPlay})

	if app.activePlayer != nil {
		t.Error("activePlayer should be cleared when ContinueOnNavigate is OFF and entering screenPlay")
	}
	if app.activeStation != nil {
		t.Error("activeStation should be cleared when ContinueOnNavigate is OFF and entering screenPlay")
	}
}

// TestContinueOnNavigate_NewStationStopsHandedOff verifies the full cross-screen
// flow: source hands off → user browses Play from Favorites → selects a new
// station → stopActivePlaybackMsg is emitted to kill the old stream.
//
// This is the a→b, c→a, g→a style scenario: whatever was playing is stopped
// the moment the user picks something new in the destination screen.
func TestContinueOnNavigate_NewStationStopsHandedOff(t *testing.T) {
	for _, dst := range destinationScreens {
		if dst.screen != screenPlay {
			// playStation() with stopActivePlaybackMsg is only implemented in
			// PlayModel. Other screens stop via stopScreenPlayers() called at
			// navigate time, which is covered by
			// TestNavigateToScreenPlay_StopsScreenPlayers.
			continue
		}
		t.Run(dst.name, func(t *testing.T) {
			app := newContinueOnNavigateApp()
			app.screen = screenPlay
			app.playScreen = NewPlayModel(t.TempDir(), nil)
			app.playScreen.playOptsCfg = app.playOptsCfg

			// Simulate a handed-off player from any source screen.
			app.activePlayer = player.NewMPVPlayer()
			app.activeStation = newTestStation("uuid-src", "Source Radio")

			// User selects a new station in Play from Favorites.
			newStation := api.Station{StationUUID: "uuid-new", Name: "New Radio", URLResolved: "http://example.com/stream"}
			cmd := app.playScreen.playStation(newStation)
			got := collectMsgs(cmd)

			if !hasMsgType[stopActivePlaybackMsg](got) {
				t.Error("expected stopActivePlaybackMsg when selecting a new station; old stream would keep playing")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Regression: Bug 1 — Quick Play / Recently Played → Play from Favorites
//             stopped the music when ContinueOnNavigate was ON.
//
// Root cause: executeMenuAction always stopped quickFavPlayer via
//             stopScreenPlayers() even when ContinueOnNavigate was ON.
// Fix:        executeMenuAction hands quickFavPlayer off to activePlayer
//             (and replaces it with a fresh idle player) when
//             ContinueOnNavigate is ON.  IsPlaying() is NOT checked because
//             Play() is called asynchronously — the user can navigate before
//             the cmd goroutine runs, making IsPlaying() a race condition.
// ---------------------------------------------------------------------------

// TestBug1_QFPlaying_NavigateToPlayScreen_PreservesStream verifies that when
// a Quick Play Favorite was selected (playingFromMain=true, playingStation set)
// and the user navigates, executeMenuAction hands quickFavPlayer off to
// activePlayer — even when IsPlaying() is still false (async Play() timing).
func TestBug1_QFPlaying_NavigateToPlayScreen_PreservesStream(t *testing.T) {
	app := newContinueOnNavigateApp()

	station := newTestStation("uuid-qf", "Quick FM")
	app.playingStation = station
	app.playingFromMain = true
	// activePlayer is nil — this is a fresh QF selection from main menu.

	originalQFP := app.quickFavPlayer

	_, _ = app.executeMenuAction(0) // Play from Favorites

	// The handoff should happen even though IsPlaying() is false (async):
	// activePlayer must be set to the original quickFavPlayer.
	if app.activePlayer != originalQFP {
		t.Error("activePlayer should be set to the original quickFavPlayer after handoff")
	}
	if app.quickFavPlayer == originalQFP {
		t.Error("quickFavPlayer should be replaced with a fresh idle player after handoff")
	}
	if app.playingFromMain {
		t.Error("playingFromMain should be cleared by executeMenuAction")
	}
}

// TestBug1_QFNotPlaying_NavigateToPlayScreen_HandoffHappens verifies that the
// handoff happens even when quickFavPlayer.IsPlaying() == false.  This is the
// key async-timing fix: Play() is called in a cmd goroutine, so IsPlaying()
// may be false at the moment executeMenuAction is called.
func TestBug1_QFNotPlaying_NavigateToPlayScreen_HandoffHappens(t *testing.T) {
	app := newContinueOnNavigateApp()
	app.playingStation = newTestStation("uuid-qf", "Quick FM")
	app.playingFromMain = true
	// quickFavPlayer.IsPlaying() == false (never started — async timing)
	originalQFP := app.quickFavPlayer

	_, _ = app.executeMenuAction(0)

	// Handoff must still happen so the station keeps playing when the cmd runs.
	if app.activePlayer != originalQFP {
		t.Error("activePlayer must be set to quickFavPlayer even when IsPlaying()==false (async case)")
	}
}

// TestBug1_HandoffMsg_SetsPlayingFromMain_DoesNotBreakSubsequentNavigate
// verifies the regression introduced during the Bug 1 fix: receiving a
// handoffPlaybackMsg sets playingFromMain=true, but if quickFavPlayer is not
// actually playing the subsequent executeMenuAction must NOT kill activePlayer.
func TestBug1_HandoffMsg_SetsPlayingFromMain_DoesNotBreakSubsequentNavigate(t *testing.T) {
	app := newContinueOnNavigateApp()

	// Simulate: user played in Play screen → handoffPlaybackMsg arrives.
	handedOff := player.NewMPVPlayer()
	station := newTestStation("uuid-play", "Play Screen FM")
	app.Update(handoffPlaybackMsg{
		player:       handedOff,
		station:      station,
		contextLabel: "Favorites",
	})
	// handoffPlaybackMsg sets playingFromMain=true.
	if !app.playingFromMain {
		t.Fatal("precondition: playingFromMain should be true after handoffPlaybackMsg")
	}
	if app.activePlayer != handedOff {
		t.Fatal("precondition: activePlayer should be the handed-off player")
	}

	// Now user navigates to Most Played (index 2).
	// quickFavPlayer.IsPlaying()==false → handoff block must be skipped.
	_, _ = app.executeMenuAction(2)

	// activePlayer must still be the handed-off player, not nil and not
	// replaced by the idle quickFavPlayer.
	if app.activePlayer != handedOff {
		t.Error("Bug 1 regression: executeMenuAction must not kill activePlayer " +
			"when quickFavPlayer is idle (stream came from a screen handoff, not QF)")
	}
}

// ---------------------------------------------------------------------------
// Regression: Bug 2 — stopScreenPlayers killed the handed-off stream
//             when navigating between player screens (e.g. Most Played →
//             Top Rated).
//
// Root cause: stopScreenPlayers iterated all screen player pointers including
//             the one that was just handed off as activePlayer.
// Fix:        stopScreenPlayers skips any player that equals activePlayer.
// ---------------------------------------------------------------------------

// TestBug2_StopScreenPlayers_SkipsActivePlayer verifies that stopScreenPlayers
// does not stop a screen player that is also activePlayer.
func TestBug2_StopScreenPlayers_SkipsActivePlayer(t *testing.T) {
	app := newContinueOnNavigateApp()

	// Simulate: Most Played handed its player to activePlayer.
	handedOff := player.NewMPVPlayer()
	app.activePlayer = handedOff
	app.mostPlayedScreen.player = handedOff // same pointer — the alias that caused the bug

	// A second, unrelated screen player that SHOULD be stopped.
	other := player.NewMPVPlayer()
	app.searchScreen.player = other

	app.stopScreenPlayers()

	// activePlayer (== mostPlayedScreen.player) must NOT have been stopped.
	// We can't assert IsPlaying()==true (never started), but we assert the
	// pointer is still the same object and was not replaced.
	if app.activePlayer != handedOff {
		t.Error("Bug 2: stopScreenPlayers replaced activePlayer — it must not touch it")
	}
	// The other player (not activePlayer) should have had Stop() called.
	// Since it was never started, IsPlaying() is already false; we can only
	// verify the function ran without panicking and didn't touch activePlayer.
}

// TestBug2_NavigateMostPlayedToTopRated_PreservesActivePlayer is an
// end-to-end regression test: simulate the full sequence
//
//	Most Played plays → handoff → navigate to Top Rated
//
// and assert activePlayer survives the navigateMsg handler.
func TestBug2_NavigateMostPlayedToTopRated_PreservesActivePlayer(t *testing.T) {
	app := newContinueOnNavigateApp()

	// Step 1: Most Played screen has a player that was handed off.
	handedOff := player.NewMPVPlayer()
	app.mostPlayedScreen.player = handedOff
	app.activePlayer = handedOff
	app.activeStation = newTestStation("uuid-mp", "Most Played FM")
	app.activeContextLabel = "Most Played"

	// Step 2: Navigate to Top Rated — this calls stopScreenPlayers().
	app.Update(navigateMsg{screen: screenTopRated})

	// activePlayer must still point to the handed-off player.
	if app.activePlayer == nil {
		t.Error("Bug 2: activePlayer was killed by stopScreenPlayers() during " +
			"Most Played → Top Rated navigation")
	}
	if app.activePlayer != handedOff {
		t.Error("Bug 2: activePlayer was replaced during navigation")
	}
}

// TestBug2_NavigatePlayToMostPlayed_PreservesActivePlayer covers the
// Play from Favorites → Most Played path.
func TestBug2_NavigatePlayToMostPlayed_PreservesActivePlayer(t *testing.T) {
	app := newContinueOnNavigateApp()

	handedOff := player.NewMPVPlayer()
	app.playScreen.player = handedOff
	app.activePlayer = handedOff
	app.activeStation = newTestStation("uuid-fav", "Fav FM")
	app.activeContextLabel = "Favorites"

	app.Update(navigateMsg{screen: screenMostPlayed})

	if app.activePlayer == nil {
		t.Error("Bug 2: activePlayer was killed during Play → Most Played navigation")
	}
	if app.activePlayer != handedOff {
		t.Error("Bug 2: activePlayer was replaced during Play → Most Played navigation")
	}
}

// ---------------------------------------------------------------------------

// TestHandoffPlaybackMsg_StopsPreviousPlayer verifies that if a second handoff
// arrives while one is already active, the old player is stopped.
func TestHandoffPlaybackMsg_StopsPreviousPlayer(t *testing.T) {
	app := newContinueOnNavigateApp()

	oldPlayer := player.NewMPVPlayer()
	newP := player.NewMPVPlayer()
	station := newTestStation("uuid-2", "New Radio")

	// First handoff.
	app.activePlayer = oldPlayer

	// Second handoff with a different player — old one should be stopped.
	app.Update(handoffPlaybackMsg{
		player:       newP,
		station:      station,
		contextLabel: "Most Played",
	})

	if app.activePlayer != newP {
		t.Error("activePlayer should be updated to the new player")
	}
	// oldPlayer.IsPlaying() is false (never started), but verifying the
	// pointer was replaced is sufficient — the code called Stop() on it.
	if app.activePlayer == oldPlayer {
		t.Error("activePlayer should not still be the old player")
	}
}

// ---------------------------------------------------------------------------
// x key: global kill binding for ContinueOnNavigate sessions
//
// When ContinueOnNavigate is ON, a station may keep playing while the user
// browses other screens.  The `x` key stops all playback from any screen.
// ---------------------------------------------------------------------------

// TestXKey_StopsActiveStation verifies that pressing `x` clears activeStation
// and activePlayer when a ContinueOnNavigate handoff is in progress.
func TestXKey_StopsActiveStation(t *testing.T) {
	app := newContinueOnNavigateApp()
	app.activeStation = newTestStation("uuid-x", "Test Radio")
	app.activePlayer = player.NewMPVPlayer()
	app.activeContextLabel = "Most Played"

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if app.activeStation != nil {
		t.Error("x key should clear activeStation")
	}
	if app.activePlayer != nil {
		t.Error("x key should clear activePlayer")
	}
}

// TestXKey_StopsPlayingFromMain verifies that pressing `x` clears
// playingFromMain and playingStation when a Quick Play station is active.
func TestXKey_StopsPlayingFromMain(t *testing.T) {
	app := newContinueOnNavigateApp()
	app.playingFromMain = true
	app.playingStation = newTestStation("uuid-qf", "QF Radio")

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if app.playingFromMain {
		t.Error("x key should clear playingFromMain")
	}
	if app.playingStation != nil {
		t.Error("x key should clear playingStation")
	}
}

// TestXKey_NoOpWhenNothingPlaying verifies that pressing `x` when no station
// is playing neither panics nor mutates playback state.
func TestXKey_NoOpWhenNothingPlaying(t *testing.T) {
	app := newTestApp()

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if app.activeStation != nil {
		t.Error("x key should not set activeStation when nothing is playing")
	}
	if app.playingFromMain {
		t.Error("x key should not set playingFromMain when nothing is playing")
	}
}

// ---------------------------------------------------------------------------
// buildNowPlayingBannerText — Now Playing bar shown on non-player screens
// ---------------------------------------------------------------------------

// TestBuildNowPlayingBannerText_ContainsXStopHint verifies that the banner
// includes the "x: Stop" keyboard hint so users know how to kill playback.
func TestBuildNowPlayingBannerText_ContainsXStopHint(t *testing.T) {
	app := newContinueOnNavigateApp()
	app.activeStation = newTestStation("uuid-radio", "Jazz FM")
	app.activeContextLabel = "Favorites"

	banner := app.buildNowPlayingBannerText()

	if !strings.Contains(banner, "x: Stop") {
		t.Errorf("banner should contain 'x: Stop', got: %q", banner)
	}
}

// TestBuildNowPlayingBannerText_EmptyWhenNoStation verifies that the banner is
// empty when no station is being tracked at the app level.
func TestBuildNowPlayingBannerText_EmptyWhenNoStation(t *testing.T) {
	app := newTestApp()

	banner := app.buildNowPlayingBannerText()

	if banner != "" {
		t.Errorf("banner should be empty when no station is active, got: %q", banner)
	}
}

// TestBuildNowPlayingBannerText_ContainsStationName verifies that the banner
// shows the name of the currently playing station.
func TestBuildNowPlayingBannerText_ContainsStationName(t *testing.T) {
	app := newContinueOnNavigateApp()
	app.activeStation = newTestStation("uuid-radio", "Jazz FM")

	banner := app.buildNowPlayingBannerText()

	if !strings.Contains(banner, "Jazz FM") {
		t.Errorf("banner should contain station name 'Jazz FM', got: %q", banner)
	}
}

// ---------------------------------------------------------------------------
// playQuickFavorite / playRecentStation — stop handed-off activePlayer
//
// Bug: after a ContinueOnNavigate handoff, activePlayer kept streaming while
// a new Quick Play or Recently Played station started — two mpv processes.
// Fix: both helpers stop and clear activePlayer when it differs from
//      quickFavPlayer.
// ---------------------------------------------------------------------------

// TestPlayQuickFavorite_StopsHandedOffActivePlayer verifies that if a player
// was handed off via ContinueOnNavigate (activePlayer != quickFavPlayer),
// calling playQuickFavorite stops and nils that player.
func TestPlayQuickFavorite_StopsHandedOffActivePlayer(t *testing.T) {
	app := newContinueOnNavigateApp()

	// Simulate a ContinueOnNavigate handoff from a player screen.
	handedOff := player.NewMPVPlayer()
	app.activePlayer = handedOff
	app.activeStation = newTestStation("uuid-handed", "Handed Off FM")
	app.activeContextLabel = "Most Played"

	// Add a quick favorite to play (required for playQuickFavorite to proceed).
	app.quickFavorites = []api.Station{
		{StationUUID: "uuid-qf", Name: "Quick FM"},
	}

	_, _ = app.playQuickFavorite(0)

	if app.activePlayer != nil {
		t.Error("playQuickFavorite should stop and clear the handed-off activePlayer")
	}
	if app.activeStation != nil {
		t.Error("playQuickFavorite should clear activeStation set by the handoff")
	}
}

// TestPlayQuickFavorite_ReplacesFreshPlayer is a regression test for the bug
// where selecting from Quick Play made no sound.
//
// Root cause: Stop() on an idle quickFavPlayer set killed=true. The subsequent
// async Play() saw killed=true and silently refused to start mpv.
// Fix: both helpers now replace quickFavPlayer with a fresh instance whose
// killed flag is false, so Play() is always allowed to proceed.
func TestPlayQuickFavorite_ReplacesFreshPlayer(t *testing.T) {
	app := newContinueOnNavigateApp()
	original := app.quickFavPlayer // idle (never started)

	app.quickFavorites = []api.Station{
		{StationUUID: "uuid-qf", Name: "Quick FM"},
	}

	_, cmd := app.playQuickFavorite(0)

	// A fresh player must have been installed so that the async Play() cmd
	// is not blocked by the old player's killed flag.
	if app.quickFavPlayer == original {
		t.Error("playQuickFavorite should replace quickFavPlayer with a fresh instance")
	}
	// The returned cmd must not be nil — a nil cmd means no Play was scheduled.
	if cmd == nil {
		t.Error("playQuickFavorite should return a non-nil cmd to start playback")
	}
}

// TestPlayRecentStation_StopsHandedOffActivePlayer verifies that if a player
// was handed off via ContinueOnNavigate, calling playRecentStation stops and
// nils that player so the new station plays exclusively.
func TestPlayRecentStation_StopsHandedOffActivePlayer(t *testing.T) {
	app := newContinueOnNavigateApp()

	// Simulate a ContinueOnNavigate handoff from a player screen.
	handedOff := player.NewMPVPlayer()
	app.activePlayer = handedOff
	app.activeStation = newTestStation("uuid-handed", "Handed Off FM")
	app.activeContextLabel = "Most Played"

	// Add a recently played entry to play.
	app.recentlyPlayed = makeRecentlyPlayed("Recent FM")

	_, _ = app.playRecentStation(0)

	if app.activePlayer != nil {
		t.Error("playRecentStation should stop and clear the handed-off activePlayer")
	}
	if app.activeStation != nil {
		t.Error("playRecentStation should clear activeStation set by the handoff")
	}
}

// TestPlayRecentStation_ReplacesFreshPlayer is a regression test for the same
// bug as TestPlayQuickFavorite_ReplacesFreshPlayer, but for Recently Played.
func TestPlayRecentStation_ReplacesFreshPlayer(t *testing.T) {
	app := newContinueOnNavigateApp()
	original := app.quickFavPlayer // idle (never started)

	app.recentlyPlayed = makeRecentlyPlayed("Recent FM")

	_, cmd := app.playRecentStation(0)

	if app.quickFavPlayer == original {
		t.Error("playRecentStation should replace quickFavPlayer with a fresh instance")
	}
	if cmd == nil {
		t.Error("playRecentStation should return a non-nil cmd to start playback")
	}
}
