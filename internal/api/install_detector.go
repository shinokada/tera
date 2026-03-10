package api

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// InstallMethod represents how tera was installed
type InstallMethod int

const (
	InstallMethodUnknown InstallMethod = iota
	InstallMethodHomebrew
	InstallMethodGo
	InstallMethodScoop
	InstallMethodWinget
	InstallMethodAPT
	InstallMethodRPM
	InstallMethodManual
)

// InstallInfo contains information about the installation
type InstallInfo struct {
	Method        InstallMethod
	UpdateCommand string
	Description   string
}

// DetectInstallMethod detects how tera was installed
func DetectInstallMethod() InstallInfo {
	// Check Homebrew (macOS/Linux)
	if checkHomebrew() {
		return InstallInfo{
			Method:        InstallMethodHomebrew,
			UpdateCommand: "brew upgrade shinokada/tera/tera",
			Description:   "Homebrew",
		}
	}

	// Check Go install
	if checkGoInstall() {
		return InstallInfo{
			Method:        InstallMethodGo,
			UpdateCommand: "go install github.com/shinokada/tera/v3/cmd/tera@latest",
			Description:   "Go Install",
		}
	}

	// Check Scoop (Windows)
	if runtime.GOOS == "windows" && checkScoop() {
		return InstallInfo{
			Method:        InstallMethodScoop,
			UpdateCommand: "scoop update tera",
			Description:   "Scoop",
		}
	}

	// Check Winget (Windows)
	if runtime.GOOS == "windows" && checkWinget() {
		return InstallInfo{
			Method:        InstallMethodWinget,
			UpdateCommand: "winget upgrade Shinokada.Tera",
			Description:   "Winget",
		}
	}

	// Check APT/DEB (Debian/Ubuntu)
	if checkAPT() {
		return InstallInfo{
			Method:        InstallMethodAPT,
			UpdateCommand: "sudo apt update && sudo apt install --only-upgrade tera",
			Description:   "APT/DEB",
		}
	}

	// Check RPM (Fedora/RHEL/CentOS)
	if checkRPM() {
		return InstallInfo{
			Method:        InstallMethodRPM,
			UpdateCommand: "sudo dnf upgrade tera",
			Description:   "RPM/DNF",
		}
	}

	// Fallback to manual
	return InstallInfo{
		Method:        InstallMethodManual,
		UpdateCommand: "", // Will show link to releases page
		Description:   "Manual/Binary",
	}
}

// runCommandWithTimeout runs a command with a 2-second timeout
func runCommandWithTimeout(name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = 100 * time.Millisecond // Give process 100ms to exit after context cancel
	return cmd.Run()
}

// checkHomebrew checks if tera was installed via Homebrew
func checkHomebrew() bool {
	// Try to run brew list tera
	if err := runCommandWithTimeout("brew", "list", "tera"); err == nil {
		return true
	}

	// Check if the binary is in a Homebrew path
	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath
	}

	// Common Homebrew paths
	homebrewPaths := []string{
		"/usr/local/Cellar",
		"/opt/homebrew/Cellar",
		"/home/linuxbrew/.linuxbrew/Cellar",
	}

	for _, path := range homebrewPaths {
		if isInDir(realPath, path) {
			return true
		}
	}

	return false
}

// isInDir checks if a given file path is inside a directory
func isInDir(filePath, dir string) bool {
	dir = filepath.Clean(dir)
	filePath = filepath.Clean(filePath)
	prefix := dir + string(os.PathSeparator)
	return filePath == dir || strings.HasPrefix(filePath, prefix)
}

// checkGoInstallPath checks if a given executable path is in the GOPATH/bin directory
func checkGoInstallPath(exePath, gopath string) bool {
	return isInDir(exePath, filepath.Join(gopath, "bin"))
}

// checkGoInstall checks if tera was installed via go install
func checkGoInstall() bool {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		gopath = filepath.Join(home, "go")
	}

	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath
	}

	// Check GOBIN first (takes precedence over GOPATH/bin)
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		if isInDir(realPath, gobin) {
			return true
		}
	}

	// Check GOPATH/bin (GOPATH can be a colon/semicolon-separated list)
	for _, gp := range filepath.SplitList(gopath) {
		if gp != "" && checkGoInstallPath(realPath, gp) {
			return true
		}
	}

	return false
}

// checkScoop checks if tera was installed via Scoop (Windows)
func checkScoop() bool {
	if err := runCommandWithTimeout("scoop", "list", "tera"); err == nil {
		return true
	}

	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath
	}

	// Check if in Scoop directory (case-insensitive for Windows)
	return strings.Contains(strings.ToLower(realPath), "\\scoop\\")
}

// checkAPT checks if tera was installed via APT/DEB
func checkAPT() bool {
	// Check if dpkg info file exists
	_, err := os.Stat("/var/lib/dpkg/info/tera.list")
	return err == nil
}

// checkWinget checks if tera was installed via Winget (Windows)
func checkWinget() bool {
	// Check winget list for tera using official package ID
	if err := runCommandWithTimeout("winget", "list", "--id", "Shinokada.Tera"); err == nil {
		return true
	}

	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath
	}

	// Check common Winget install locations (case-insensitive for Windows)
	wingetPaths := []string{
		"\\program files\\",
		"\\appdata\\local\\microsoft\\winget\\packages\\",
	}

	lowerPath := strings.ToLower(realPath)
	for _, path := range wingetPaths {
		if strings.Contains(lowerPath, path) {
			return true
		}
	}

	return false
}

// checkRPM checks if tera was installed via RPM
func checkRPM() bool {
	return runCommandWithTimeout("rpm", "-q", "tera") == nil
}
