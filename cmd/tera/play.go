package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/player"
	"github.com/shinokada/tera/v3/internal/storage"
)

// handlePlay is the entry point for `tera play ...`.
// --duration is pre-extracted from rawArgs before flag.FlagSet.Parse so that
// it is recognised regardless of where it appears relative to the source
// argument (e.g. `tera play fav --duration 30m` and
// `tera play --duration 30m fav` both work). Go's standard flag parser stops
// at the first non-flag positional, so without this pre-scan the flag would
// be silently ignored when placed after the source name.
func handlePlay(rawArgs []string) {
	// Pre-scan for --duration / --duration=VALUE so it works in any position.
	durationStr, filteredArgs := extractDurationFlag(rawArgs)

	playCmd := flag.NewFlagSet("play", flag.ExitOnError)
	playCmd.Usage = printPlayHelp
	if err := playCmd.Parse(filteredArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	args := playCmd.Args()

	if len(args) == 0 {
		printPlayHelp()
		return
	}

	// Parse --duration
	var dur time.Duration
	if durationStr != "" {
		var err error
		dur, err = time.ParseDuration(durationStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid duration %q (use e.g. 30s, 10m, 1h, 1h30m)\n", durationStr)
			os.Exit(1)
		}
		if dur <= 0 {
			fmt.Fprintf(os.Stderr, "Error: duration must be positive\n")
			os.Exit(1)
		}
	}

	// Reject unrecognised option-like tokens before dispatching to sub-parsers.
	// Without this guard, flags like --help fall through and are mis-handled
	// (e.g. treated as a list name by parseFavArgs or silently ignored by parseNArg).
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			fmt.Fprintf(os.Stderr, "Error: unknown flag %q\n\n", a)
			printPlayHelp()
			os.Exit(1)
		}
	}

	switch args[0] {
	case "--help", "-h":
		printPlayHelp()
	case "favorites", "fav":
		listName, n := parseFavArgs(args[1:])
		handlePlayFavorites(listName, n, dur)
	case "recent", "rec":
		n := parseNArg(args[1:])
		handlePlayRecent(n, dur)
	case "top-rated", "top":
		n := parseNArg(args[1:])
		handlePlayTopRated(n, dur)
	case "most-played", "most":
		n := parseNArg(args[1:])
		handlePlayMostPlayed(n, dur)
	case "lucky":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tera play lucky <keyword>")
			os.Exit(1)
		}
		keyword := joinLuckyKeyword(args[1:])
		handlePlayLucky(keyword, dur)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown play source %q\n\n", args[0])
		printPlayHelp()
		os.Exit(1)
	}
}

// extractDurationFlag scans rawArgs for --duration VALUE or --duration=VALUE,
// removes those tokens, and returns the value string and the remaining args.
// This allows --duration to appear anywhere (before or after the source name)
// without tripping Go's flag parser, which stops at the first positional arg.
func extractDurationFlag(rawArgs []string) (durationStr string, rest []string) {
	rest = make([]string, 0, len(rawArgs))
	for i := 0; i < len(rawArgs); i++ {
		arg := rawArgs[i]
		// --duration=VALUE form
		if strings.HasPrefix(arg, "--duration=") {
			durationStr = strings.TrimPrefix(arg, "--duration=")
			if durationStr == "" {
				fmt.Fprintln(os.Stderr, "Error: --duration requires a value (e.g. --duration 30m)")
				os.Exit(1)
			}
			continue
		}
		// --duration VALUE form
		if arg == "--duration" || arg == "-duration" {
			if i+1 < len(rawArgs) && !strings.HasPrefix(rawArgs[i+1], "-") {
				i++
				durationStr = rawArgs[i]
			} else {
				fmt.Fprintln(os.Stderr, "Error: --duration requires a value (e.g. --duration 30m)")
				os.Exit(1)
			}
			continue
		}
		rest = append(rest, arg)
	}
	return
}

// parseFavArgs extracts [list-name] and [n] from the args following "fav".
// Both are optional:
//
//	(none)        → My-favorites, 1
//	jazz          → jazz, 1
//	jazz 3        → jazz, 3
//	3             → My-favorites, 3   (numeric-only first arg is treated as n)
func parseFavArgs(args []string) (listName string, n int) {
	listName = "My-favorites"
	n = 1

	if len(args) == 0 {
		return
	}

	// If the first arg is a pure integer, treat it as n (list defaults)
	if v, err := strconv.Atoi(args[0]); err == nil {
		n = v
		return
	}

	// First arg is a list name
	listName = args[0]

	if len(args) >= 2 {
		if v, err := strconv.Atoi(args[1]); err == nil {
			n = v
		}
	}
	return
}

// parseNArg extracts an optional [n] from args, defaulting to 1.
func parseNArg(args []string) int {
	if len(args) == 0 {
		return 1
	}
	v, err := strconv.Atoi(args[0])
	if err != nil || v < 1 {
		return 1
	}
	return v
}

// joinLuckyKeyword joins keyword args into a single space-separated string.
// This allows multi-word keywords: tera play lucky smooth jazz → "smooth jazz"
func joinLuckyKeyword(args []string) string {
	return strings.Join(args, " ")
}

// truncate shortens s to at most maxLen runes, appending "…" if truncated.
// Returns an empty string if maxLen is 0 or negative.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

// dataDir returns the path to the tera data directory.
func dataDir() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not determine config directory: %w", err)
	}
	return filepath.Join(cfgDir, "tera", "data"), nil
}

// newMetadataManager creates a MetadataManager using the standard data path.
func newMetadataManager() (*storage.MetadataManager, error) {
	dir, err := dataDir()
	if err != nil {
		return nil, err
	}
	return storage.NewMetadataManager(dir)
}

// newRatingsManager creates a RatingsManager using the standard data path.
func newRatingsManager() (*storage.RatingsManager, error) {
	dir, err := dataDir()
	if err != nil {
		return nil, err
	}
	return storage.NewRatingsManager(dir)
}

// favoritesDir returns the path to the favorites directory, honouring the
// TERA_FAVORITE_PATH environment variable override (same logic as the TUI).
func favoritesDir() (string, error) {
	if override := os.Getenv("TERA_FAVORITE_PATH"); override != "" {
		return override, nil
	}
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "favorites"), nil
}

// -----------------------------------------------------------------
// runPlayback is the shared core: start mpv, print the status line,
// block until Ctrl+C or --duration fires, then clean up.
//
// meta may be nil (e.g. when called from handlePlayFavorites which has no
// pre-opened manager). When non-nil it is reused so only one MetadataManager
// instance is open at a time; Close() is always called before returning.
// -----------------------------------------------------------------
func runPlayback(station *api.Station, contextLabel string, dur time.Duration, meta *storage.MetadataManager) {
	// If no manager was passed in, open one now (non-fatal on failure).
	if meta == nil {
		var metaErr error
		meta, metaErr = newMetadataManager()
		if metaErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not open metadata store: %v\n", metaErr)
		}
	}
	if meta != nil {
		defer func() {
			if err := meta.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not flush play stats: %v\n", err)
			}
		}()
	}

	p := player.NewMPVPlayer()
	if meta != nil {
		p.SetMetadataManager(meta)
	}

	if err := p.Play(station); err != nil {
		if meta != nil {
			if closeErr := meta.Close(); closeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not flush play stats: %v\n", closeErr)
			}
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print status line
	name := truncate(station.Name, 40)
	if dur > 0 {
		fmt.Printf("▶ Playing: %s  [%s]  (stops in %s · Ctrl+C to stop early)\n",
			name, contextLabel, dur)
	} else {
		fmt.Printf("▶ Playing: %s  [%s]  (Ctrl+C to stop)\n",
			name, contextLabel)
	}

	// Set up signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	var stopMsg string
	if dur > 0 {
		timer := time.NewTimer(dur)
		defer timer.Stop()
		select {
		case <-sigChan:
			stopMsg = "Stopped."
		case <-timer.C:
			stopMsg = "Stopped (duration reached)."
		case <-p.Done():
			fmt.Println("\nStopped (stream ended).")
			return // mpv already cleaned up; skip p.Stop()
		}
	} else {
		select {
		case <-sigChan:
			stopMsg = "Stopped."
		case <-p.Done():
			fmt.Println("\nStopped (stream ended).")
			return // mpv already cleaned up; skip p.Stop()
		}
	}

	if err := p.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: could not stop playback cleanly: %v\n", err)
	}
	fmt.Printf("\n%s\n", stopMsg)
}

// -----------------------------------------------------------------
// handlePlayFavorites: tera play fav [list-name] [n]
// -----------------------------------------------------------------
func handlePlayFavorites(listName string, n int, dur time.Duration) {
	favDir, err := favoritesDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	store := storage.NewStorage(favDir)
	list, err := store.LoadList(context.Background(), listName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: list %q not found\n", listName)
		} else {
			fmt.Fprintf(os.Stderr, "Error: could not load list %q: %v\n", listName, err)
		}
		os.Exit(1)
	}

	total := len(list.Stations)
	if total == 0 {
		fmt.Fprintf(os.Stderr, "Error: list %q is empty\n", listName)
		os.Exit(1)
	}

	if n < 1 || n > total {
		fmt.Fprintf(os.Stderr, "Error: %q has %d station(s). Please choose 1–%d.\n",
			listName, total, total)
		os.Exit(1)
	}

	station := list.Stations[n-1]
	label := fmt.Sprintf("%s · item %d of %d", listName, n, total)
	// handlePlayFavorites has no pre-opened MetadataManager, pass nil so
	// runPlayback opens one itself.
	runPlayback(&station, label, dur, nil)
}

// -----------------------------------------------------------------
// handlePlayRecent: tera play recent [n]
// -----------------------------------------------------------------
func handlePlayRecent(n int, dur time.Duration) {
	meta, err := newMetadataManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	// No defer Close() here — on the happy path runPlayback takes ownership
	// and closes meta. Early exits close explicitly to stop the saveLoop goroutine.

	results := meta.GetRecentlyPlayed(0) // 0 = no limit
	if len(results) == 0 {
		_ = meta.Close()
		fmt.Fprintln(os.Stderr, "Error: no recently played stations found")
		os.Exit(1)
	}

	if n < 1 || n > len(results) {
		_ = meta.Close()
		fmt.Fprintf(os.Stderr, "Error: only %d station(s) in recently played. Please choose 1–%d.\n",
			len(results), len(results))
		os.Exit(1)
	}

	station := results[n-1].Station
	if station.Name == "" {
		station.Name = "[unknown]"
	}
	label := fmt.Sprintf("recently played · #%d", n)
	// Pass the already-open manager so runPlayback reuses it rather than
	// opening a second instance against the same file.
	runPlayback(&station, label, dur, meta)
}

// -----------------------------------------------------------------
// handlePlayTopRated: tera play top [n]
// -----------------------------------------------------------------
func handlePlayTopRated(n int, dur time.Duration) {
	ratings, err := newRatingsManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	// No defer Close() here — early exits close explicitly to stop the
	// saveLoop goroutine; on the happy path we close before entering runPlayback.

	results := ratings.GetTopRated(0) // 0 = no limit
	if len(results) == 0 {
		_ = ratings.Close()
		fmt.Fprintln(os.Stderr, "Error: no rated stations found")
		os.Exit(1)
	}

	if n < 1 || n > len(results) {
		_ = ratings.Close()
		fmt.Fprintf(os.Stderr, "Error: only %d rated station(s). Please choose 1–%d.\n",
			len(results), len(results))
		os.Exit(1)
	}

	item := results[n-1]
	station := item.Station
	if station.Name == "" {
		station.Name = "[unknown]"
	}
	stars := storage.RenderStarsCompact(item.Rating.Rating, true)
	label := fmt.Sprintf("top rated · %s", stars)
	// Close before runPlayback: RatingsManager has no data left to write here,
	// and runPlayback opens its own MetadataManager for play stats.
	_ = ratings.Close()
	runPlayback(&station, label, dur, nil)
}

// -----------------------------------------------------------------
// handlePlayMostPlayed: tera play most [n]
// -----------------------------------------------------------------
func handlePlayMostPlayed(n int, dur time.Duration) {
	meta, err := newMetadataManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	// No defer Close() here — on the happy path runPlayback takes ownership
	// and closes meta. Early exits close explicitly to stop the saveLoop goroutine.

	results := meta.GetTopPlayed(0) // 0 = no limit
	if len(results) == 0 {
		_ = meta.Close()
		fmt.Fprintln(os.Stderr, "Error: no play history found")
		os.Exit(1)
	}

	if n < 1 || n > len(results) {
		_ = meta.Close()
		fmt.Fprintf(os.Stderr, "Error: only %d station(s) in play history. Please choose 1–%d.\n",
			len(results), len(results))
		os.Exit(1)
	}

	item := results[n-1]
	station := item.Station
	if station.Name == "" {
		station.Name = "[unknown]"
	}
	label := fmt.Sprintf("most played · %d plays", item.Metadata.PlayCount)
	// Pass the already-open manager so runPlayback reuses it.
	runPlayback(&station, label, dur, meta)
}

// -----------------------------------------------------------------
// handlePlayLucky: tera play lucky <keyword>
// -----------------------------------------------------------------
func handlePlayLucky(keyword string, dur time.Duration) {
	fmt.Printf("Searching for %q...\n", keyword)

	// Fixed timeout for Radio Browser API calls.
	const apiTimeout = 15 * time.Second
	timeout := apiTimeout

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := api.NewClient()
	params := api.SearchParams{
		Tag:        keyword,
		Name:       keyword,
		Order:      "votes",
		Reverse:    true,
		Limit:      100,
		HideBroken: true,
	}

	stations, err := client.SearchAdvanced(ctx, params)
	if err != nil {
		if ctx.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error: could not reach Radio Browser API (timeout)\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error: could not reach Radio Browser API: %v\n", err)
		}
		os.Exit(1)
	}

	// Filter out stations with no resolved URL
	valid := make([]api.Station, 0, len(stations))
	for _, s := range stations {
		if s.URLResolved != "" {
			valid = append(valid, s)
		}
	}

	if len(valid) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no stations found for %q\n", keyword)
		os.Exit(1)
	}

	// Pick a random station (matching TUI "I Feel Lucky" behaviour)
	//nolint:gosec // not used for cryptographic purposes
	station := valid[rand.Intn(len(valid))]
	label := fmt.Sprintf("lucky · %q", keyword)
	runPlayback(&station, label, dur, nil)
}

// -----------------------------------------------------------------
// printPlayHelp prints usage for `tera play`.
// -----------------------------------------------------------------
func printPlayHelp() {
	fmt.Print(`TERA Play Commands

Usage: tera play <source> [args] [--duration <duration>]

Sources:
  favorites, fav      [list-name] [n]   Play nth station from a favorites list
  recent, rec         [n]               Play the nth most recently played station
  top-rated, top      [n]               Play the nth highest-rated station
  most-played, most   [n]               Play the nth most-played station
  lucky               <keyword ...>     Play a random station matching keyword(s)

Options:
  --duration  Stop after duration (e.g. 30s, 10m, 1h, 1h30m)

Defaults:
  list-name  My-favorites
  n          1 (first item)

Examples:
  tera play fav
  tera play fav jazz 3
  tera play recent 2 / tera play rec 2
  tera play top
  tera play most-played 3
  tera play lucky smooth jazz
  tera play fav --duration 30m
  tera play lucky ambient --duration 1h
`)
}
