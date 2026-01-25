# Fixing Dependency Issues

## Problem
The tests are failing because `go.sum` is missing entries for some transitive dependencies:
- `github.com/atotto/clipboard` (used by bubbles/textinput)
- `github.com/sahilm/fuzzy` (used by bubbles/list)

## Solution

Run this in your terminal:

```bash
cd /Users/shinichiokada/Terminal-Tools/tera

# Add the missing dependencies
go get github.com/atotto/clipboard
go get github.com/sahilm/fuzzy

# Clean up
go mod tidy

# Now test
go test ./internal/ui -v -run Play
```

Or use the script:

```bash
chmod +x fix_and_test.sh
./fix_and_test.sh
```

## Why This Happens

The `bubbles` library has these as dependencies, but they weren't pulled into `go.sum` yet because:
1. We hadn't imported/used the specific bubbles components that need them
2. Go's module system only adds entries when actually compiling code that uses them

## After Fixing

You should see output like:
```text
=== RUN   TestNewPlayModel
--- PASS: TestNewPlayModel (0.00s)
=== RUN   TestGetAvailableLists
--- PASS: TestGetAvailableLists (0.00s)
=== RUN   TestPlayModel_Update_ListsLoaded
--- PASS: TestPlayModel_Update_ListsLoaded (0.00s)
...
PASS
ok      github.com/shinokada/tera/internal/ui    0.XXXs
```
