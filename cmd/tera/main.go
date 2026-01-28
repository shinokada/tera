package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/theme"
	"github.com/shinokada/tera/internal/ui"
)

func main() {
	// Handle CLI arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "theme":
			handleThemeCommand()
			return
		case "--help", "-h":
			printHelp()
			return
		case "--version", "-v":
			fmt.Println("TERA v0.1.0")
			return
		}
	}

	// Initialize theme before starting UI
	if _, err := theme.Load(); err != nil {
		fmt.Printf("Warning: Could not load theme: %v\n", err)
	}

	p := tea.NewProgram(ui.NewApp(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
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
		configPath, _ := theme.GetConfigPath()
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
		fmt.Println("Edit this file to customize colors and padding.")

	default:
		printThemeHelp()
	}
}

func printThemeHelp() {
	fmt.Println(`TERA Theme Commands

Usage: tera theme <command>

Commands:
  reset    Reset theme to default values
  path     Show path to theme config file
  edit     Show theme config file location for editing

The theme config file uses YAML format and includes:
  - colors: primary, secondary, highlight, error, success, muted, text
  - padding: page margins, list item spacing, box padding

Color values can be:
  - ANSI color numbers (0-255)
  - Hex colors (#FF5733 or #F53)

Example: Edit ~/.config/tera/theme.yaml to customize your theme.`)
}

func printHelp() {
	fmt.Println(`TERA - Terminal Radio Player

Usage: tera [command]

Commands:
  theme    Manage theme settings (reset, path, edit)

Options:
  -h, --help     Show this help message
  -v, --version  Show version

Run without arguments to start the interactive radio player.`)
}
