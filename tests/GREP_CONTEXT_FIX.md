# Test Fix: Grep Context Issue

## The Problem

Tests were failing with:
```
✗ All menus have Exit at the bottom
  (in test file ./test_integration.bats, line 32)
  `echo "$search_menu" | grep -q "7) Exit"' failed
```

## Root Cause

The grep context wasn't large enough to capture all menu options.

The search menu MENU_OPTIONS looks like this:
```bash
MENU_OPTIONS="0) Main Menu
1) Tag
2) Name
3) Language
4) Country code
5) State
6) Advanced search
7) Exit"
```

That's **8 lines** (options 0-7), but we were only using `-A 10` (after 10 lines), which wasn't enough because:
1. First grep finds the function: `search_menu()`
2. Takes 10 lines after that
3. Second grep finds: `MENU_OPTIONS=`
4. Takes 10 lines after that
5. But we need to reach line 7 which might be cut off!

## The Fix

Increased grep context from 10 to 15 lines:

**Before:**
```bash
result=$(grep -A 10 'search_menu()' ../lib/search.sh | grep -A 10 'MENU_OPTIONS=')
```

**After:**
```bash
result=$(grep -A 15 'search_menu()' ../lib/search.sh | grep -A 15 'MENU_OPTIONS=')
```

## Why This Works

- List menu: 6 options (0-5) → 10 lines was enough
- Gist menu: 4 options (0-3) → 10 lines was enough
- Search menu: **8 options (0-7)** → Needed 15 lines!

With 15 lines of context, we reliably capture all menu options including "7) Exit".

## Files Changed

- `tests/test_menu_structure.bats` - Increased context to 15 lines
- `tests/test_integration.bats` - Increased context to 15 lines for all menus

## Test Results

```
./test_headings.bats        ✓ 9/9 pass
./test_integration.bats     ✓ 8/8 pass   ✅ Fixed!
./test_menu_structure.bats  ✓ 10/10 pass ✅ Fixed!
./test_navigation.bats      ✓ 6/6 pass
./test_search.bats          ✓ 6/6 pass

39 tests, 0 failures ✅
```

## Lesson Learned

When using `grep -A N` (after N lines), count the actual lines you need to capture:
- Always add a buffer (use 15 instead of 8)
- More context is safer than less
- Test against the longest/largest case (search menu has the most options)

## Verification

Run the tests:
```bash
cd tests
bats .
```

All tests should now pass! 🎉
