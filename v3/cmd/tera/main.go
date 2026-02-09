// Package main provides the TERA terminal-based internet radio player.
//
// TERA is a keyboard-driven radio player powered by Radio Browser that supports
// searching stations by name, tag, language, country, or state. Features include:
//   - Search and browse thousands of internet radio stations
//   - Organize favorites into custom lists
//   - Quick play shortcuts for instant station access
//   - Gist sync for backing up favorites via GitHub
//   - Customizable themes via YAML configuration
//
// Requirements:
//   - mpv must be installed for audio playback
//
// Basic usage:
//
//	tera                  # Start the application
//	tera theme path       # Show theme config location
//	tera theme reset      # Reset theme to defaults
//	tera --version        # Show version
//	tera --help           # Show help
//
// For complete documentation, visit: https://tera.codewithshin.com/
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/config"
	"github.com/shinokada/tera/v3/internal/storage"
	"github.com/shinokada/tera/v3/internal/theme"
	"github.com/shinokada/tera/v3/internal/ui"
)

// Version is set at build time via -ldflags "-X main.version=v1.0.0"
var version = "dev"

func getVersion() string {
	// If version was set at build time, use it
	if version != "dev" {
		return version
	}
	// Otherwise try to get version from Go module info (works with go install)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func main() {
	// Handle CLI arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "theme":
			handleThemeCommand()
			return
		case "config":
			handleConfigCommand()
			return
		case "--help", "-h":
			printHelp()
			return
		case "--version", "-v":
			fmt.Printf("TERA %s\n", getVersion())
			return
		}
	}

	// Check and migrate v2 config if needed
	migrated, err := storage.CheckAndMigrateV2Config(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Config migration failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "Starting with default configuration...")
	} else if migrated {
		fmt.Println("🔄 Migrated TERA configuration from v2 to v3")
		if configPath, err := config.GetConfigPath(); err == nil {
			fmt.Printf("✓ Config unified → %s\n", configPath)
		} else {
			fmt.Println("✓ Config unified into config.yaml")
		}
		fmt.Println("✓ Old configs backed up with timestamp")
		fmt.Println()
	}

	// Initialize theme before starting UI
	if _, err := theme.Load(); err != nil {
		fmt.Printf("Warning: Could not load theme: %v\n", err)
	}

	// Set version in UI package for About screen
	ui.Version = getVersion()

	app := ui.NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Set up graceful shutdown handler for SIGINT (Ctrl+C) and SIGTERM
	// This ensures proper cleanup even when signals bypass Bubble Tea's key handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		// Clean up resources (stop audio players, close files, etc.)
		app.Cleanup()
		// Send quit to Bubble Tea program to restore terminal
		p.Send(tea.Quit())
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleConfigCommand() {
	if len(os.Args) < 3 {
		printConfigHelp()
		return
	}

	switch os.Args[2] {
	case "path":
		configPath, err := config.GetConfigPath()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(configPath)

	case "reset":
		if err := config.Reset(); err != nil {
			fmt.Printf("Error resetting config: %v\n", err)
			os.Exit(1)
		}
		configPath, _ := config.GetConfigPath()
		fmt.Println("✓ Configuration reset to defaults")
		fmt.Printf("  Config file: %s\n", configPath)

	case "validate":
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
		if err := cfg.Validate(); err != nil {
			fmt.Printf("⚠️  Configuration has issues:\n%v\n", err)
			fmt.Println("\nConfig has been auto-corrected. Run 'tera config reset' to restore defaults.")
		} else {
			fmt.Println("✓ Configuration is valid")
		}

	case "migrate":
		configDir, err := os.UserConfigDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		v2ConfigDir := filepath.Join(configDir, "tera")

		if !config.HasV2Config(v2ConfigDir) {
			fmt.Println("No v2 configuration found.")
			return
		}

		detected := config.DetectV2Config(v2ConfigDir)
		fmt.Println("V2 configuration detected:")
		// Iterate in deterministic order for consistent output
		for _, file := range []string{"theme.yaml", "appearance_config.yaml", "connection_config.yaml", "shuffle.yaml"} {
			if detected[file] {
				fmt.Printf("  ✓ %s\n", file)
			}
		}

		if len(os.Args) > 3 && os.Args[3] == "--force" {
			migrated, err := storage.CheckAndMigrateV2Config(true)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			if migrated {
				fmt.Println("\n✓ Migration complete!")
			}
		} else {
			fmt.Println("\nRun TERA normally to auto-migrate, or use 'tera config migrate --force' to migrate now.")
		}

	default:
		printConfigHelp()
	}
}

func handleThemeCommand() {
	if len(os.Args) < 3 {
		printThemeHelp()
		return
	}

	switch os.Args[2] {
	case "reset":
		if err := theme.Reset(); err != nil {
			fmt.Printf("Error resetting theme: %v\n", err)
			os.Exit(1)
		}
		configPath, _ := config.GetConfigPath()
		fmt.Println("✓ Theme reset to defaults")
		fmt.Printf("  Config file: %s\n", configPath)

	case "path":
		configPath, err := theme.GetConfigPath()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(configPath)

	case "edit":
		configPath, err := theme.GetConfigPath()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		// Ensure config file exists
		if _, err := theme.Load(); err != nil {
			fmt.Printf("Error loading theme: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Theme config file: %s\n", configPath)
		fmt.Println("Edit the 'ui.theme' section to customize colors and padding.")

	case "export":
		// Export current theme as standalone theme.yaml
		outputPath := "theme.yaml"
		if len(os.Args) > 3 {
			outputPath = os.Args[3]
		}
		if err := theme.ExportLegacyThemeFile(outputPath); err != nil {
			fmt.Printf("Error exporting theme: %v\n", err)
			os.Exit(1)
		}
		absPath, _ := filepath.Abs(outputPath)
		fmt.Printf("✓ Theme exported to %s\n", absPath)

	default:
		printThemeHelp()
	}
}

func printConfigHelp() {
	fmt.Println(`TERA Configuration Commands

Usage: tera config <command>

Commands:
  path       Show path to config file
  reset      Reset all settings to defaults
  validate   Check configuration for errors
  migrate    Check for v2 config and show migration status

The config file (config.yaml) contains:
  - player: playback settings (volume, buffer)
  - ui: theme colors and appearance
  - network: connection and streaming
  - shuffle: shuffle mode behavior

Example: Run 'tera config path' to find your config file location.`)
}

func printThemeHelp() {
	fmt.Println(`TERA Theme Commands

Usage: tera theme <command>

Commands:
  reset    Reset theme to default values
  path     Show path to config file
  edit     Show config file location for theme editing
  export   Export current theme as standalone theme.yaml

The theme is now part of the unified config.yaml file under 'ui.theme'.
Theme settings include:
  - colors: primary, secondary, highlight, error, success, muted, text
  - padding: page margins, list item spacing, box padding

Color values can be:
  - ANSI color numbers (0-255)
  - Hex colors (#FF5733 or #F53)

Example: Run 'tera config path' to find your config file, then edit the ui.theme section.`)
}

func printHelp() {
	fmt.Println(`TERA - Terminal Radio Player

Usage: tera [command]

Commands:
  theme    Manage theme settings (reset, path, edit, export)
  config   Manage configuration (path, reset, validate, migrate)

Options:
  -h, --help     Show this help message
  -v, --version  Show version

Run without arguments to start the interactive radio player.

Version 3.0 introduces unified configuration:
  - All settings in one config.yaml (run 'tera config path' to find it)
  - Automatic migration from v2
  - Better organization and validation`)
}
