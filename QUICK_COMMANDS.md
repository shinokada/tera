# Quick Commands - Shuffle Mode Implementation

## Build and Test Commands

```bash
# Navigate to project directory
cd /Users/shinichiokada/Terminal-Tools/tera

# Clean build
make clean && make lint && make build

# Run all tests
make test

# Run only shuffle tests
go test ./internal/ui -v -run "TestLuckyShuffle"

# Run only lucky tests
go test ./internal/ui -v -run "TestLucky"

# Run with coverage
make coverage

# Run the application
./tera
```

## Git Commands (when ready to commit)

```bash
# Check status
git status

# Add files
git add internal/shuffle/manager.go
git add internal/ui/lucky.go
git add internal/ui/lucky_test.go
git add README.md

# Commit
git commit -m "feat: Add shuffle mode to I Feel Lucky

- Toggle shuffle mode with 't' key in I Feel Lucky
- Auto-advance timer with configurable intervals (1-15 min)
- Station history with backward navigation
- Pause/resume timer controls
- Full playback controls during shuffle
- Configurable via Settings → Shuffle Settings
- Settings persist to ~/.config/tera/shuffle.yaml

Closes #XXX" # Replace XXX with issue number if applicable

# Push
git push origin main  # or your branch name
```

## Testing Workflow

```bash
# 1. Build
make clean && make build

# 2. Quick test - run automated tests
make test

# 3. Manual test - launch app
./tera

# Navigate in TERA:
# - Press 4 (I Feel Lucky)
# - Press 't' (Enable shuffle)
# - Type 'jazz' and Enter
# - Test controls: n, b, p, h, f, s, v
# - Press Esc to exit

# 4. Verify configuration
cat ~/.config/tera/shuffle.yaml
ls -la ~/.config/tera/

# 5. Test settings
./tera
# - Press 6 (Settings)
# - Press 2 (Shuffle Settings)
# - Test all settings options
# - Verify changes persist
```

## Cleanup Commands (if needed)

```bash
# Remove build artifacts
make clean

# Remove configuration (for fresh test)
rm ~/.config/tera/shuffle.yaml

# Remove all TERA config (full reset)
rm -rf ~/.config/tera/

# Kill any stuck mpv processes
pkill mpv
# or on macOS
killall mpv
```

## Documentation Files Created

All in `/Users/shinichiokada/Terminal-Tools/tera/`:

```bash
# View documentation
cat README.md | grep -A 100 "Shuffle Mode"
cat SHUFFLE_QUICK_REFERENCE.md
cat SHUFFLE_TESTING_CHECKLIST.md
cat SHUFFLE_IMPLEMENTATION_COMPLETE.md
cat SHUFFLE_MODE_SUMMARY.md
cat SHUFFLE_FINAL_SUMMARY.md
cat SHUFFLE_BUGFIX.md

# Or open in your editor
code README.md
code SHUFFLE_QUICK_REFERENCE.md
```

## Quick Validation

```bash
# Verify all files exist
ls -la internal/shuffle/manager.go
ls -la internal/ui/lucky.go
ls -la internal/ui/lucky_test.go
ls -la README.md

# Check for compilation errors
go build -o tera cmd/tera/main.go
echo "Build status: $?"  # Should output 0

# Run linter
golangci-lint run ./...
echo "Lint status: $?"   # Should output 0

# Count lines added
git diff --stat internal/shuffle/manager.go
git diff --stat internal/ui/lucky.go
git diff --stat internal/ui/lucky_test.go
```

## Troubleshooting

```bash
# If build fails, check for missing dependencies
go mod tidy
go mod download

# If tests fail, run with verbose output
go test -v ./internal/ui -run TestLuckyShuffle

# If linting fails, try auto-fix
make lint-fix

# If shuffle doesn't start, check config exists
ls -la ~/.config/tera/shuffle.yaml
cat ~/.config/tera/shuffle.yaml

# If timer doesn't work, verify interval is set
grep interval ~/.config/tera/shuffle.yaml

# If no sound, verify mpv is installed
which mpv
mpv --version
```

## Performance Testing

```bash
# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Profile memory usage
go test -memprofile=mem.prof ./internal/ui
go tool pprof mem.prof

# Profile CPU usage
go test -cpuprofile=cpu.prof ./internal/ui
go tool pprof cpu.prof
```

## Release Preparation

```bash
# 1. Verify everything works
make clean && make lint && make build && make test

# 2. Update version (if using versioning)
# Edit cmd/tera/main.go or version.go

# 3. Update CHANGELOG.md
# Add new shuffle mode feature

# 4. Tag release (example)
git tag -a v1.x.x -m "Add shuffle mode feature"
git push origin v1.x.x

# 5. Create GitHub release
# Use release notes from SHUFFLE_FINAL_SUMMARY.md
```

## Useful Aliases (optional)

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
alias tera-build='cd /Users/shinichiokada/Terminal-Tools/tera && make clean && make build'
alias tera-test='cd /Users/shinichiokada/Terminal-Tools/tera && make test'
alias tera-run='cd /Users/shinichiokada/Terminal-Tools/tera && ./tera'
alias tera-dev='cd /Users/shinichiokada/Terminal-Tools/tera && make clean && make build && ./tera'
```

---

## Summary

The shuffle mode is complete and ready! 🎉

**Quick start:**
```bash
cd /Users/shinichiokada/Terminal-Tools/tera
make clean && make build
./tera
# Press 4, then 't', then type 'jazz' and Enter
```

All documentation is in the project directory. Use the testing checklist to verify everything works!
