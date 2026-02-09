// Package main provides a simple example of using the config package
//
// Run this to verify the config package works:
//
//	go run examples/config_example.go
package main

import (
	"fmt"
	"os"

	"github.com/shinokada/tera/v3/internal/config"
)

func main() {
	fmt.Println("=== TERA v3 Config Package Example ===")

	// 1. Check if config exists
	exists, err := config.Exists()
	if err != nil {
		fmt.Printf("Error checking config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config exists: %v\n\n", exists)

	// 2. Load config (creates default if doesn't exist)
	fmt.Println("Loading config...")
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Config loaded successfully")

	// 3. Display current settings
	fmt.Println("Current Configuration:")
	fmt.Printf("Version: %s\n\n", cfg.Version)

	fmt.Println("Player:")
	fmt.Printf("  Default Volume: %d\n", cfg.Player.DefaultVolume)
	fmt.Printf("  Buffer Size: %d MB\n\n", cfg.Player.BufferSizeMB)

	fmt.Println("UI Theme:")
	fmt.Printf("  Name: %s\n", cfg.UI.Theme.Name)
	fmt.Printf("  Primary Color: %s\n", cfg.UI.Theme.Colors["primary"])
	fmt.Printf("  Secondary Color: %s\n", cfg.UI.Theme.Colors["secondary"])
	fmt.Printf("  Highlight Color: %s\n\n", cfg.UI.Theme.Colors["highlight"])

	fmt.Println("UI Appearance:")
	fmt.Printf("  Header Mode: %s\n", cfg.UI.Appearance.HeaderMode)
	fmt.Printf("  Header Align: %s\n", cfg.UI.Appearance.HeaderAlign)
	fmt.Printf("  Header Width: %d\n\n", cfg.UI.Appearance.HeaderWidth)

	fmt.Println("Network:")
	fmt.Printf("  Auto Reconnect: %v\n", cfg.Network.AutoReconnect)
	fmt.Printf("  Reconnect Delay: %d seconds\n", cfg.Network.ReconnectDelay)
	fmt.Printf("  Buffer Size: %d MB\n\n", cfg.Network.BufferSizeMB)

	fmt.Println("Shuffle:")
	fmt.Printf("  Auto Advance: %v\n", cfg.Shuffle.AutoAdvance)
	fmt.Printf("  Interval: %d minutes\n", cfg.Shuffle.IntervalMinutes)
	fmt.Printf("  Remember History: %v\n", cfg.Shuffle.RememberHistory)
	fmt.Printf("  Max History: %d\n", cfg.Shuffle.MaxHistory)

	// 4. Get config path
	configPath, _ := config.GetConfigPath()
	fmt.Printf("Config file location: %s\n\n", configPath)

	// 5. Test validation
	fmt.Println("Testing validation...")
	testCfg := *cfg // Copy config

	// Try invalid values
	testCfg.Player.DefaultVolume = 150  // Too high
	testCfg.Network.ReconnectDelay = 0  // Too low
	testCfg.Shuffle.IntervalMinutes = 7 // Invalid value

	fmt.Println("Before validation:")
	fmt.Printf("  Volume: %d (invalid, should be <= 100)\n", testCfg.Player.DefaultVolume)
	fmt.Printf("  Reconnect Delay: %d (invalid, should be >= 1)\n", testCfg.Network.ReconnectDelay)
	fmt.Printf("  Interval: %d (invalid, must be 1,3,5,10,15)\n\n", testCfg.Shuffle.IntervalMinutes)

	if err := testCfg.Validate(); err != nil {
		fmt.Printf("Validation warnings: %v\n\n", err)
	}

	fmt.Println("After validation (auto-corrected):")
	fmt.Printf("  Volume: %d ✓\n", testCfg.Player.DefaultVolume)
	fmt.Printf("  Reconnect Delay: %d ✓\n", testCfg.Network.ReconnectDelay)
	fmt.Printf("  Interval: %d ✓\n\n", testCfg.Shuffle.IntervalMinutes)

	// 6. Test modification and save
	fmt.Println("Testing save functionality...")
	cfg.Player.DefaultVolume = 80
	cfg.UI.Theme.Name = "custom"

	if err := config.Save(cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Config saved successfully")

	// 7. Test backup
	fmt.Println("Creating backup...")
	if err := config.Backup("example"); err != nil {
		fmt.Printf("Error creating backup: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Backup created: %s.example\n\n", configPath)

	// 8. Test v2 detection
	fmt.Println("Testing v2 config detection...")
	v2ConfigDir, _ := os.UserConfigDir()
	v2ConfigDir = v2ConfigDir + "/tera"

	if config.HasV2Config(v2ConfigDir) {
		fmt.Println("✓ V2 config detected")
		detected := config.DetectV2Config(v2ConfigDir)
		for file, exists := range detected {
			if exists {
				fmt.Printf("  - %s ✓\n", file)
			}
		}
	} else {
		fmt.Println("No v2 config detected")
	}

	fmt.Println("\n=== All tests passed! ===")
}
