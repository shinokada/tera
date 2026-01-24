# TERA Test Suite

This directory contains automated tests for the TERA radio player application.

## Test Framework

The tests use **BATS** (Bash Automated Testing System), a TAP-compliant testing framework for Bash scripts.

## Installing BATS

### macOS (via Homebrew)
```bash
brew install bats-core
```

### Linux (Ubuntu/Debian)
```bash
sudo apt-get install bats
```

### Manual Installation
```bash
git clone https://github.com/bats-core/bats-core.git
cd bats-core
sudo ./install.sh /usr/local
```

## Running Tests

### Run all tests
```bash
cd tests
bats .
```

### Run specific test file
```bash
bats test_menu_structure.bats
bats test_gist_crud.bats
bats test_gist_menu_integration.bats
```

### Run with verbose output
```bash
bats -t .
```

## Test Files

### `test_gist_crud.bats` ⭐ NEW
Tests Gist CRUD (Create, Read, Update, Delete) operations:
- Gist metadata initialization
- Creating and saving gist metadata
- Retrieving all gists and specific gists by ID
- Updating gist descriptions and timestamps
- Deleting gists from metadata
- Gist count operations
- Data integrity across multiple operations
- ISO 8601 timestamp format validation
- Edge cases (empty metadata, nonexistent gists)

### `test_gist_menu_integration.bats` ⭐ NEW
Integration tests for Gist Menu functionality:
- Gist metadata file creation
- Metadata persistence after gist creation
- List view showing correct count
- Gist selection and recovery
- Deletion from both GitHub and local metadata
- Gist menu display with counts
- URL and description storage
- Multiple gists coexistence
- Order maintenance
- Empty and corrupted file handling

### `test_station_names.bats`
Tests station name trimming and alphabetical sorting:
- Verifies whitespace trimming from station names
- Tests alphabetical sorting (case-insensitive)
- Validates jq gsub pattern for trimming
- Tests edge cases (special characters, long names, duplicates)
- Performance tests with large lists
- Real-world station name handling

### `test_menu_structure.bats`
Tests menu structure and consistency:
- Verifies all menus have Main Menu at position 0
- Checks Exit is at the bottom
- Validates menu option ordering

### `test_headings.bats`
Tests page headings:
- Verifies all pages have appropriate headings
- Checks heading format (Title Case)
- Validates use of cyanprint for headings
- Tests FZF header flags

### `test_navigation.bats`
Tests navigation and ESC key behavior:
- Verifies ESC returns to menus (not quit)
- Tests Main Menu option handling
- Validates empty input handling

### `test_search.bats`
Tests search functionality:
- Verifies "Searching..." message cleanup
- Tests Main Menu option in search results
- Validates station number adjustment
- Checks search header consistency

### `test_integration.bats`
Integration tests for overall consistency:
- Tests consistent menu conventions across all menus
- Verifies Main Menu option in all interactive selections
- Validates FZF prompt consistency
- Checks for redundant text removal

### `test_gist_improvements.bats`
Tests for gist menu improvements:
- Recovery from gist improvements
- Gist URL handling

### `test_duplicates.bats`
Tests duplicate detection:
- Verifies duplicate station detection
- Tests duplicate warning messages

## Expected Test Results

All tests should pass if the recent changes are correctly implemented:
- ✓ Gist metadata CRUD operations work correctly (NEW)
- ✓ Gist metadata persists across operations (NEW)
- ✓ Gist timestamps use ISO 8601 format (NEW)
- ✓ Multiple gists can coexist (NEW)
- ✓ Gist deletion removes from both GitHub and metadata (NEW)
- ✓ Gist recovery supports selection from saved gists (NEW)
- ✓ Station names have whitespace trimmed
- ✓ Stations displayed in alphabetical order
- ✓ Case-insensitive sorting works correctly
- ✓ Main Menu at position 0 in all menus
- ✓ << Main Menu >> option in all FZF selections
- ✓ Clear headings on all pages
- ✓ ESC key returns to menu (not quit)
- ✓ "Searching..." message cleaned up after search
- ✓ Consistent prompts and navigation

## Manual Testing

For quick manual verification of station name improvements:
```bash
cd tests
chmod +x manual_test_station_improvements.sh
./manual_test_station_improvements.sh
```

This script tests:
- Alphabetical sorting of your actual favorite lists
- Whitespace detection in station names
- jq trimming pattern functionality

## Testing Gist Features

### Unit Tests
The gist CRUD tests verify:
- Metadata file creation and initialization
- Adding, retrieving, updating, and deleting gists
- Data integrity and persistence
- Timestamp formatting
- Edge cases and error handling

### Integration Tests
The gist menu integration tests verify:
- End-to-end gist workflows
- Menu interactions
- GitHub API integration points (mocked)
- User experience flows

### Manual Testing for Gist Features

To manually test gist CRUD operations:

1. **Test Create:**
   ```bash
   ./tera
   # Select: 6) Gist
   # Select: 1) Create a gist
   # Verify: Gist is created and saved to metadata
   ```

2. **Test My Gists:**
   ```bash
   ./tera
   # Select: 6) Gist
   # Select: 2) My Gists
   # Verify: All created gists are listed with timestamps
   ```

3. **Test Recover:**
   ```bash
   ./tera
   # Select: 6) Gist
   # Select: 3) Recover favorites from a gist
   # Try both: selecting a number and entering a URL
   # Verify: Lists are downloaded correctly
   ```

4. **Test Delete:**
   ```bash
   ./tera
   # Select: 6) Gist
   # Select: 4) Delete a gist
   # Verify: Gist is removed from both GitHub and metadata
   ```

## Continuous Integration

You can integrate these tests into your CI/CD pipeline by adding:

```yaml
# Example GitHub Actions
- name: Run BATS tests
  run: |
    brew install bats-core
    cd tests
    bats .
```

## Writing New Tests

When adding new features, create corresponding tests:

1. Add test file: `test_feature_name.bats`
2. Use descriptive test names with `@test "description"`
3. Follow existing patterns for setup/teardown
4. Test both positive and negative cases
5. Include edge case testing
6. Verify error handling

### Example Test Structure

```bash
#!/usr/bin/env bats

setup() {
    # Set up test environment
    export TEST_DIR="$BATS_TEST_DIRNAME/test_temp"
    mkdir -p "$TEST_DIR"
    
    # Source required libraries
    source "$BATS_TEST_DIRNAME/../lib/your_lib.sh"
}

teardown() {
    # Clean up
    rm -rf "$TEST_DIR"
}

@test "feature does what it should" {
    # Arrange
    # Act
    # Assert
    [ condition ]
}
```

## Test Coverage

Current test coverage includes:

### Core Features
- ✅ Menu structure and navigation
- ✅ Station search functionality
- ✅ List management
- ✅ Station name handling
- ✅ Duplicate detection

### Gist Features (NEW)
- ✅ Gist metadata storage (CRUD)
- ✅ Create gist workflow
- ✅ List saved gists
- ✅ Recover from gist (with selection)
- ✅ Delete gist workflow
- ✅ Data persistence
- ✅ Timestamp handling
- ✅ Error cases

### Integration
- ✅ End-to-end workflows
- ✅ Cross-feature interactions
- ✅ Menu flow consistency

## Troubleshooting

### Tests failing?
- Make sure you're running tests from the `tests/` directory
- Check that all lib files are in `../lib/`
- Verify BATS is properly installed: `bats --version`
- Ensure jq is installed: `jq --version`

### Gist tests failing?
- Check that `gist_storage.sh` is in `../lib/`
- Verify JSON formatting in test data
- Check file permissions in test directories
- Ensure date command works with ISO 8601 format

### Path issues?
Tests assume they're run from the `tests/` directory. Adjust paths if running from elsewhere.

### JSON parsing errors?
- Verify jq is installed and working
- Check JSON syntax in test data
- Ensure proper escaping in test strings

## Test Development Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Always clean up in teardown
3. **Descriptive Names**: Use clear, descriptive test names
4. **Edge Cases**: Test both success and failure paths
5. **Documentation**: Comment complex test logic
6. **Fast**: Keep tests quick (mock expensive operations)
7. **Deterministic**: Tests should always produce same results

## Future Test Areas

Potential areas for additional test coverage:
- GitHub API error handling
- Network timeout scenarios
- Concurrent gist operations
- Large gist file handling
- UTF-8 and special character handling in gist content
- Token validation and expiration
