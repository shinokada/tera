package api

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectInstallMethod(t *testing.T) {
	// Basic smoke test - should not panic
	info := DetectInstallMethod()
	
	// Should return a valid method
	if info.Method < InstallMethodUnknown || info.Method > InstallMethodManual {
		t.Errorf("Invalid install method: %d", info.Method)
	}
	
	// Should have a description
	if info.Description == "" {
		t.Error("Install info should have a description")
	}
	
	// If not manual/unknown, should have an update command
	if info.Method != InstallMethodManual && info.Method != InstallMethodUnknown {
		if info.UpdateCommand == "" {
			t.Errorf("Install method %s should have an update command", info.Description)
		}
	}
}

func TestCheckGoInstall(t *testing.T) {
	tests := []struct {
		name     string
		gopath   string
		exePath  string
		expected bool
	}{
		{
			name:     "binary in GOPATH/bin",
			gopath:   "/home/user/go",
			exePath:  "/home/user/go/bin/tera",
			expected: true,
		},
		{
			name:     "binary not in GOPATH",
			gopath:   "/home/user/go",
			exePath:  "/usr/local/bin/tera",
			expected: false,
		},
		{
			name:     "false positive - bin2 directory",
			gopath:   "/home/user/go",
			exePath:  "/home/user/go/bin2/tera",
			expected: false,
		},
		{
			name:     "false positive - binary directory",
			gopath:   "/home/user/go",
			exePath:  "/home/user/go/binary/tera",
			expected: false,
		},
		{
			name:     "exact match to bin directory",
			gopath:   "/home/user/go",
			exePath:  "/home/user/go/bin",
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkGoInstallPath(tt.exePath, tt.gopath)
			if result != tt.expected {
				t.Errorf("checkGoInstallPath(%q, %q) = %v, want %v",
					tt.exePath, tt.gopath, result, tt.expected)
			}
		})
	}
	
	// Also test the actual function doesn't crash
	_ = checkGoInstall()
}

func TestCheckHomebrew(t *testing.T) {
	// Smoke test - should not panic
	_ = checkHomebrew()
	
	// Platform-specific test
	if runtime.GOOS == "darwin" {
		// On macOS, Homebrew is common
		result := checkHomebrew()
		t.Logf("Homebrew detection on macOS: %v", result)
	}
}

func TestCheckScoop(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Scoop is Windows-only")
	}
	
	// Smoke test
	_ = checkScoop()
}

func TestCheckWinget(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Winget is Windows-only")
	}
	
	// Smoke test
	_ = checkWinget()
}

func TestCheckAPT(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("APT is Linux-only")
	}
	
	result := checkAPT()
	// Check if the function runs without error
	t.Logf("APT detection: %v", result)
}

func TestCheckRPM(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("RPM is Linux-only")
	}
	
	result := checkRPM()
	// Check if the function runs without error
	t.Logf("RPM detection: %v", result)
}

func TestInstallMethodStrings(t *testing.T) {
	tests := []struct {
		method      InstallMethod
		shouldHaveCommand bool
	}{
		{InstallMethodHomebrew, true},
		{InstallMethodGo, true},
		{InstallMethodScoop, true},
		{InstallMethodWinget, true},
		{InstallMethodAPT, true},
		{InstallMethodRPM, true},
		{InstallMethodManual, false},
		{InstallMethodUnknown, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.method.String(), func(t *testing.T) {
			// Create an InstallInfo with this method
			var info InstallInfo
			info.Method = tt.method
			
			// Set appropriate values based on method
			switch tt.method {
			case InstallMethodHomebrew:
				info.UpdateCommand = "brew upgrade shinokada/tera/tera"
				info.Description = "Homebrew"
			case InstallMethodGo:
				info.UpdateCommand = "go install github.com/shinokada/tera/cmd/tera@latest"
				info.Description = "Go Install"
			case InstallMethodScoop:
				info.UpdateCommand = "scoop update tera"
				info.Description = "Scoop"
			case InstallMethodWinget:
				info.UpdateCommand = "winget upgrade tera"
				info.Description = "Winget"
			case InstallMethodAPT:
				info.UpdateCommand = "sudo apt update && sudo apt upgrade tera"
				info.Description = "APT/DEB"
			case InstallMethodRPM:
				info.UpdateCommand = "sudo dnf upgrade tera"
				info.Description = "RPM/DNF"
			case InstallMethodManual:
				info.Description = "Manual/Binary"
			case InstallMethodUnknown:
				info.Description = "Unknown"
			}
			
			// Verify update command presence matches expectation
			hasCommand := info.UpdateCommand != ""
			if hasCommand != tt.shouldHaveCommand {
				t.Errorf("Method %s: expected command=%v, got command=%v",
					tt.method.String(), tt.shouldHaveCommand, hasCommand)
			}
		})
	}
}

func TestDetectInstallMethodRealistic(t *testing.T) {
	// Get the actual detection result
	info := DetectInstallMethod()
	
	t.Logf("Detected install method: %s", info.Description)
	t.Logf("Update command: %s", info.UpdateCommand)
	
	// Verify we got a valid result
	if info.Description == "" {
		t.Error("Description should not be empty")
	}
	
	// If method is not manual/unknown, should have a command
	if info.Method != InstallMethodManual && info.Method != InstallMethodUnknown {
		if info.UpdateCommand == "" {
			t.Error("Non-manual install should have an update command")
		}
	}
	
	// Verify the update command is reasonable for the method
	switch info.Method {
	case InstallMethodHomebrew:
		if info.UpdateCommand != "brew upgrade shinokada/tera/tera" {
			t.Errorf("Unexpected Homebrew command: %s", info.UpdateCommand)
		}
	case InstallMethodGo:
		if info.UpdateCommand != "go install github.com/shinokada/tera/cmd/tera@latest" {
			t.Errorf("Unexpected Go command: %s", info.UpdateCommand)
		}
	}
}

// Helper method to convert InstallMethod to string for testing
func (i InstallMethod) String() string {
	switch i {
	case InstallMethodHomebrew:
		return "Homebrew"
	case InstallMethodGo:
		return "Go"
	case InstallMethodScoop:
		return "Scoop"
	case InstallMethodWinget:
		return "Winget"
	case InstallMethodAPT:
		return "APT"
	case InstallMethodRPM:
		return "RPM"
	case InstallMethodManual:
		return "Manual"
	case InstallMethodUnknown:
		return "Unknown"
	default:
		return "Invalid"
	}
}

func TestExecutablePathDetection(t *testing.T) {
	// Test that we can get the executable path
	exePath, err := os.Executable()
	if err != nil {
		t.Fatalf("Failed to get executable path: %v", err)
	}
	
	t.Logf("Executable path: %s", exePath)
	
	// Try to resolve symlinks
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath
	}
	
	t.Logf("Real path: %s", realPath)
	
	// Verify path is absolute
	if !filepath.IsAbs(realPath) {
		t.Error("Real path should be absolute")
	}
}
