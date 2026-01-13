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
bats test_headings.bats
bats test_navigation.bats
bats test_search.bats
bats test_integration.bats
```

### Run with verbose output
```bash
bats -t .
```

## Test Files

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

## Expected Test Results

All tests should pass if the recent changes are correctly implemented:
- ✓ Main Menu at position 0 in all menus
- ✓ << Main Menu >> option in all FZF selections
- ✓ Clear headings on all pages
- ✓ ESC key returns to menu (not quit)
- ✓ "Searching..." message cleaned up after search
- ✓ Consistent prompts and navigation

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

## Troubleshooting

### Tests failing?
- Make sure you're running tests from the `tests/` directory
- Check that all lib files are in `../lib/`
- Verify BATS is properly installed: `bats --version`

### Path issues?
Tests assume they're run from the `tests/` directory. Adjust paths if running from elsewhere.
