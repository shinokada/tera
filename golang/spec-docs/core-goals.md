# Core goals

- Convert Bash to Golang 
- Better UX/UI

## Benefits

- Native TUI libraries - Libraries like bubbletea + bubbles could be great for what I'm building. The entire menu system, keyboard handling, and search would be 10x easier.
- Single binary - Distribute one executable, no runtime dependencies. Users run brew install lim and it just works.
- Trivial cross-compilation for Windows/Mac/Linux
- Fast startup and low memory usage
- Built-in JSON - Native JSON marshaling/unmarshaling. No jq dependency, faster parsing.
- Proper error handling - Go's error handling makes the code clearer and more robust.
- Great testing - Go's testing framework is excellent, and you can easily mock Homebrew calls.
- Performance - Instant startup, responsive UI, fast JSON parsing.
- Goroutines - Easy to show progress bars while running brew install in background.
- Type safety - Catch bugs at compile time instead of runtime during user testing.

## What TERA does

- ✅ TUI with menus/lists (Bubble Tea excels)
- ✅ HTTP API calls (Go's net/http is excellent)
- ✅ JSON parsing (Go's encoding/json is great)
- ✅ Subprocess control (mpv player)
- ✅ File I/O (config, favorites)
- ✅ Cross-platform (Go's strength)

## Repository Strategy: Branch based approach

- ✅ Clean separation of Bash and Go codebases
- ✅ Each branch has clean root directory (no bash/ or go/ subdirectories)
- ✅ Independent releases and versioning
- ✅ Clear migration path
- ✅ Preserves all stars, issues, community in one repo
- ✅ Eventually deprecate Bash cleanly