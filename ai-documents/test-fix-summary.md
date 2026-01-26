# Test Fix Summary

## Issue
Two failing tests in `internal/ui/play_test.go`:
- Line 32: Expected error when no lists available, got nil
- Line 58: Expected 3 lists, got 4

## Root Cause
`NewPlayModel()` automatically creates:
1. The favorites directory (`os.MkdirAll`)
2. `My-favorites.json` file (if it doesn't exist)

This means tests always start with at least one list file.

## Solution
Updated `TestGetAvailableLists` to account for auto-created `My-favorites.json`:

### Before
```go
// Test with no files
model := NewPlayModel(tmpDir)
lists, err := model.getAvailableLists()
if err == nil {
    t.Error("Expected error when no lists available, got nil")
}

// Later...
expectedCount := 3
if len(lists) != expectedCount {
    t.Errorf("Expected %d lists, got %d", expectedCount, len(lists))
}
```

### After
```go
// Test with no files - but NewPlayModel creates My-favorites.json automatically
model := NewPlayModel(tmpDir)
lists, err := model.getAvailableLists()
// Should find My-favorites.json that was auto-created
if err != nil {
    t.Errorf("Expected no error with auto-created My-favorites, got: %v", err)
}
if len(lists) != 1 {
    t.Errorf("Expected 1 list (My-favorites), got %d", len(lists))
}
if lists[0] != "My-favorites" {
    t.Errorf("Expected 'My-favorites', got '%s'", lists[0])
}

// Later...
// My-favorites.json was auto-created, plus 3 test files = 4 total
expectedCount := 4
if len(lists) != expectedCount {
    t.Errorf("Expected %d lists (including auto-created My-favorites), got %d", expectedCount, len(lists))
}
```

## Tests Now Verify
1. `My-favorites.json` is auto-created on initialization
2. List count includes the auto-created file
3. The auto-created list has the correct name
