package api

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallMethod represents how tera was installed
type InstallMethod int

const (
	InstallMethodUnknown InstallMethod = iota
	InstallMethodHomebrew
	InstallMethodGo
	InstallMethodScoop
	InstallMethodAPT
	InstallMethodRPM
	InstallMethodManual
)

// InstallInfo contains information about the installation
type InstallInfo struct {
	Method       InstallMethod
	UpdateCommand string
	Description   string
}

// DetectInstallMethod detects how tera was installed
func DetectInstallMethod() InstallInfo {
	// Check Homebrew (macOS/Linux)
	if checkHomebrew() {
		return InstallInfo{
			Method:       InstallMethodHomebrew,
			UpdateCommand: "brew upgrade tera",
			Description:   "Homebrew",
		}
	}

	// Check Go install
	if checkGoInstall() {
		return InstallInfo{
			Method:       InstallMethodGo,
			UpdateCommand: "go install github.com/shinokada/tera/cmd/tera@latest",
			Description:   "Go Install",
		}
	}

	// Check Scoop (Windows)
	if runtime.GOOS == "windows" && checkScoop() {
		return InstallInfo{
			Method:       InstallMethodScoop,
			UpdateCommand: "scoop update tera",
			Description:   "Scoop",
		}
	}

	// Check APT/DEB (Debian/Ubuntu)
	if checkAPT() {
		return InstallInfo{
			Method:       InstallMethodAPT,
			UpdateCommand: "sudo apt update && sudo apt upgrade tera",
			Description:   "APT/DEB",
		}
	}

	// Check RPM (Fedora/RHEL/CentOS)
	if checkRPM() {
		return InstallInfo{
			Method:       InstallMethodRPM,
			UpdateCommand: "sudo dnf upgrade tera",
			Description:   "RPM/DNF",
		}
	}

	// Fallback to manual
	return InstallInfo{
		Method:       InstallMethodManual,
		UpdateCommand: "", // Will show link to releases page
		Description:   "Manual/Binary",
	}
}

// checkHomebrew checks if tera was installed via Homebrew
func checkHomebrew() bool {
	// Try to run brew list tera
	cmd := exec.Command("brew", "list", "tera")
	if err := cmd.Run(); err == nil {
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
		if strings.HasPrefix(realPath, path) {
			return true
		}
	}

	return false
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

	// Check if binary is in GOPATH/bin
	gopathBin := filepath.Join(gopath, "bin")
	return strings.HasPrefix(realPath, gopathBin)
}

// checkScoop checks if tera was installed via Scoop (Windows)
func checkScoop() bool {
	cmd := exec.Command("scoop", "list", "tera")
	if err := cmd.Run(); err == nil {
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

	// Check if in Scoop directory
	return strings.Contains(realPath, "\\scoop\\")
}

// checkAPT checks if tera was installed via APT/DEB
func checkAPT() bool {
	// Check if dpkg info file exists
	_, err := os.Stat("/var/lib/dpkg/info/tera.list")
	return err == nil
}

// checkRPM checks if tera was installed via RPM
func checkRPM() bool {
	cmd := exec.Command("rpm", "-q", "tera")
	return cmd.Run() == nil
}
