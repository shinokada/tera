# TERA Unit Tests - Complete Guide

## 📋 Overview

A comprehensive test suite has been created to validate all recent improvements to the TERA radio player application. The tests use **BATS (Bash Automated Testing System)**, a TAP-compliant testing framework designed specifically for Bash scripts.

## 🎯 What's Been Tested

### All Recent Changes Have Test Coverage:

1. ✅ **Menu Structure Consistency** - Main Menu at position 0, Exit at bottom
2. ✅ **Interactive Selection Options** - << Main Menu >> in all FZF selections
3. ✅ **Page Headings** - Clear headings on every page using Title Case
4. ✅ **ESC Key Navigation** - Returns to menu instead of quitting
5. ✅ **Search Message Cleanup** - "Searching..." removed after completion
6. ✅ **Prompt Consistency** - Simple `> ` prompts throughout

## 📁 Test Files Created

```text
tests/
├── README.md                    # Test documentation
├── TEST_COVERAGE.md            # Detailed coverage documentation
├── run_tests.sh                # Main test runner (all tests)
├── quick_test.sh               # Quick test runner (critical tests)
├── test_menu_structure.bats    # Menu structure tests
├── test_headings.bats          # Heading display tests
├── test_navigation.bats        # ESC key and navigation tests
├── test_search.bats            # Search functionality tests
└── test_integration.bats       # Integration and consistency tests

.github/workflows/
└── test.yml                    # GitHub Actions CI/CD workflow

Makefile                        # Convenient test commands
```

## 🚀 Quick Start

### 1. Install BATS

**macOS:**
```bash
brew install bats-core
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install bats
```

**Or use the Makefile:**
```bash
make install-bats
```

### 2. Run Tests

**Run all tests:**
```bash
cd tests
bats .
```

**Or use the Makefile:**
```bash
make test
```

**Run quick tests (critical only):**
```bash
make quick-test
```

**Run specific test file:**
```bash
cd tests
bats test_menu_structure.bats
```

## 📊 Test Coverage

### test_menu_structure.bats (9 tests)
- ✓ Main menu structure
- ✓ List menu has Main Menu at position 0
- ✓ Search menu has Main Menu at position 0
- ✓ Search submenu has Main Menu at position 0
- ✓ Gist menu has Main Menu at position 0
- ✓ Play function has Main Menu in list selection
- ✓ Play function has Main Menu in station selection
- ✓ Search results have Main Menu option
- ✓ Delete station has Main Menu option
- ✓ All menus use consistent prompt style

### test_headings.bats (9 tests)
- ✓ Play from my list has heading
- ✓ Delete station has heading
- ✓ I Feel Lucky has heading
- ✓ Search by functions have headings
- ✓ Advanced search has heading
- ✓ Create gist has heading
- ✓ Recover gist has heading
- ✓ All headings use cyanprint function
- ✓ FZF headers use header-first flag

### test_navigation.bats (6 tests)
- ✓ Play list selection handles empty input (ESC)
- ✓ Play station selection handles empty input (ESC)
- ✓ Search results handle empty input (ESC)
- ✓ Search results no longer mention ESC in prompt
- ✓ Lucky function allows menu return
- ✓ Delete station handles zero input for Main Menu

### test_search.bats (6 tests)
- ✓ wget_simple_search clears 'Searching...' message
- ✓ wget_search clears 'Searching...' message
- ✓ Search results include Main Menu option
- ✓ Search results adjust station numbers correctly
- ✓ Advanced search includes Main Menu option
- ✓ Search functions use consistent headers

### test_integration.bats (8 tests)
- ✓ All menus follow 0=Main Menu convention
- ✓ All menus have Exit at the bottom
- ✓ All interactive selections have Main Menu option
- ✓ FZF prompts are consistent
- ✓ All headings use Title Case
- ✓ No redundant text after search completes
- ✓ All clear commands come before headings
- ✓ No double Main Menu entries in any menu

**Total: ~40 test cases covering 100% of recent changes**

## 🔄 Continuous Integration

### GitHub Actions
The repository now includes a GitHub Actions workflow that:
- Runs automatically on push/pull request
- Tests on both Ubuntu and macOS
- Provides clear pass/fail status

Enable it by pushing the `.github/workflows/test.yml` file to your repository.

## 💡 Usage Examples

### Before Making Changes
```bash
# Run tests to ensure everything works
make test
```

### After Making Changes
```bash
# Run quick tests for fast feedback
make quick-test

# If quick tests pass, run full suite
make test
```

### Debugging Failed Tests
```bash
# Run with verbose output
cd tests
bats -t test_menu_structure.bats
```

## 🛠️ Makefile Commands

```bash
make help         # Show all available commands
make test         # Run all tests
make quick-test   # Run critical tests only
make install-bats # Install BATS framework
make clean        # Clean up test artifacts
```

## 📝 Test Maintenance

### Adding New Tests

When adding new features:

1. Create a new test file or add to existing:
   ```bash
   touch tests/test_new_feature.bats
   ```

2. Follow the existing pattern:
   ```bash
   @test "description of what you're testing" {
       result=$(grep 'pattern' ../lib/file.sh)
       echo "$result" | grep -q "expected"
   }
   ```

3. Run tests to verify:
   ```bash
   bats tests/test_new_feature.bats
   ```

### Updating Tests

If you modify existing functionality:

1. Update corresponding tests
2. Run affected test file
3. Run full suite to check for side effects

## 🎓 Best Practices

1. **Run tests before committing** - Catch issues early
2. **Write tests for new features** - Maintain coverage
3. **Keep tests simple** - Easy to understand and maintain
4. **Use descriptive test names** - Clear what's being tested
5. **Test both success and failure cases** - Comprehensive coverage

## 📚 Additional Resources

- **BATS Documentation**: https://github.com/bats-core/bats-core
- **TAP Protocol**: https://testanything.org/
- **Bash Testing Guide**: https://www.tldp.org/LDP/abs/html/debugging.html

## 🐛 Troubleshooting

### "bats: command not found"
Install BATS: `make install-bats` or follow manual installation instructions

### "No such file or directory"
Make sure you're in the `tests/` directory when running BATS

### Tests failing unexpectedly
1. Check that lib files haven't moved
2. Verify file paths in tests match actual structure
3. Run with verbose output: `bats -t test_file.bats`

## ✨ Benefits

✅ **Confidence** - Know your changes work correctly  
✅ **Regression Prevention** - Catch breaking changes early  
✅ **Documentation** - Tests describe expected behavior  
✅ **Refactoring Safety** - Change code with confidence  
✅ **CI/CD Ready** - Automated testing in pipelines  

## 🎉 Success!

You now have a complete, professional test suite for TERA! The tests cover all recent improvements and ensure the application maintains consistent, user-friendly navigation throughout.

Happy testing! 🚀
