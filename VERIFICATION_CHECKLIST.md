# Bug Fixes Verification Checklist

## Pre-Testing Setup

- [ ] Read `BUG_FIXES_COMPLETE.md`
- [ ] Read `FIXES_SUMMARY.md`  
- [ ] Review `VISUAL_FIXES_GUIDE.md`
- [ ] Run: `chmod +x build_and_verify.sh`
- [ ] Run: `./build_and_verify.sh`
- [ ] Confirm: Build successful
- [ ] Confirm: MPV installed

---

## Issue 1: Station Stops on Quit

### Test 1A: Quit from Play Screen
- [ ] Run `./tera`
- [ ] Press `1` (Play from Favorites)
- [ ] Select a list
- [ ] Select a station
- [ ] Wait for audio to start
- [ ] Press `q`
- [ ] **Expected:** Audio stops immediately
- [ ] **Expected:** Returns to station list

### Test 1B: Quit from Search Play
- [ ] Run `./tera`
- [ ] Press `2` (Search)
- [ ] Search for stations
- [ ] Select and play one
- [ ] Press `q`
- [ ] **Expected:** Audio stops immediately
- [ ] **Expected:** Save prompt appears (unless already saved)

### Test 1C: Force Quit with Ctrl+C
- [ ] Run `./tera`
- [ ] Play any station
- [ ] Press `Ctrl+C`
- [ ] **Expected:** Audio stops immediately
- [ ] **Expected:** Clean exit

### Test 1D: Verify No Zombie Processes
- [ ] After each quit test above
- [ ] Run: `ps aux | grep mpv`
- [ ] **Expected:** No MPV processes running
- [ ] **Expected:** Empty result (except grep itself)

**Result:** ⬜ Pass ⬜ Fail  
**Notes:**

---

## Issue 2 & 5: Screen Heights

### Test 2A: Search Menu Height
- [ ] Run `./tera` in standard terminal (80x24 minimum)
- [ ] Press `2` (Search)
- [ ] **Expected:** See all 6 options without scrolling:
  - [ ] 1. Search by Tag
  - [ ] 2. Search by Name
  - [ ] 3. Search by Language
  - [ ] 4. Search by Country Code
  - [ ] 5. Search by State
  - [ ] 6. Advanced Search
- [ ] **Expected:** Help text visible at bottom

### Test 2B: Play Screen List Height
- [ ] Run `./tera`
- [ ] Press `1` (Play from Favorites)
- [ ] **Expected:** List of favorites uses most of screen
- [ ] **Expected:** Can see multiple lists without scrolling

### Test 2C: Play Screen Station Height  
- [ ] From Play screen
- [ ] Select a list with many stations
- [ ] **Expected:** Station list uses most of screen
- [ ] **Expected:** Can see many stations

### Test 2D: Search Results Height
- [ ] Search for stations with many results
- [ ] **Expected:** Results list uses most of screen
- [ ] **Expected:** Can see 10+ stations

### Test 2E: Window Resize
- [ ] On any screen with a list
- [ ] Make terminal larger
- [ ] **Expected:** List grows
- [ ] Make terminal smaller  
- [ ] **Expected:** List shrinks (minimum 5 lines)
- [ ] **Expected:** Still usable at minimum

**Result:** ⬜ Pass ⬜ Fail  
**Notes:**

---

## Issue 3: Save Prompt After Search Play

### Test 3A: New Station Save Prompt
- [ ] Run `./tera`
- [ ] Press `2` (Search)
- [ ] Search by tag: "jazz"
- [ ] Select a station NOT in My-favorites
- [ ] Press `1` (Play)
- [ ] Wait for audio
- [ ] Press `q`
- [ ] **Expected:** See save prompt:
  ```
  💾 Save Station?
  
  Did you enjoy this station?
  [Station Name]
  
  1) ⭐ Add to Quick Favorites
  2) Return to search results
  
  y/1: Yes • n/2/Esc: No
  ```

### Test 3B: Save - Yes Option
- [ ] From save prompt
- [ ] Press `y` or `1`
- [ ] **Expected:** See success message
- [ ] **Expected:** Station added to My-favorites
- [ ] **Expected:** Return to search results

### Test 3C: Save - No Option
- [ ] Search and play another station
- [ ] Press `q` at save prompt
- [ ] Press `n` or `2` or `Esc`
- [ ] **Expected:** Return to results
- [ ] **Expected:** Station NOT saved

### Test 3D: Already Saved Station
- [ ] Play same station from Test 3B again
- [ ] Press `q`
- [ ] **Expected:** See message "Already in Quick Favorites"
- [ ] **Expected:** No save prompt
- [ ] **Expected:** Return to results

### Test 3E: Verify Saved Station
- [ ] Return to main menu
- [ ] Check Quick Favorites (if available)
- [ ] **Expected:** Station from Test 3B is there
- [ ] Play it from Quick Favorites
- [ ] **Expected:** No save prompt (already saved)

**Result:** ⬜ Pass ⬜ Fail  
**Notes:**

---

## Issue 4: Filter Count Updates

### Test 4A: Basic Filter
- [ ] Search for stations (100+ results ideal)
- [ ] Note total count shown
- [ ] Press `/` (activate filter)
- [ ] **Expected:** Filter input appears
- [ ] Type: "rock"
- [ ] **Expected:** See status bar at bottom
- [ ] **Expected:** Shows "x/y items" where y = total
- [ ] **Expected:** x changes as you type

### Test 4B: Filter Refinement
- [ ] Continue typing to narrow results
- [ ] **Expected:** Count decreases
- [ ] Delete characters
- [ ] **Expected:** Count increases
- [ ] Clear filter (delete all)
- [ ] **Expected:** Back to full count

### Test 4C: No Matches Filter
- [ ] Type nonsense: "zzzzzz"
- [ ] **Expected:** Shows "0/[total] items"
- [ ] **Expected:** Empty list message

**Result:** ⬜ Pass ⬜ Fail  
**Notes:**

---

## Additional Verification

### Terminal Size Tests
- [ ] Test on 80x24 terminal (minimum)
- [ ] Test on 120x40 terminal (medium)
- [ ] Test on 200x60 terminal (large)
- [ ] All sizes remain usable

### Edge Cases
- [ ] Quit while filter is active
- [ ] Resize during save prompt
- [ ] Multiple quick quit/play cycles
- [ ] Very long station names
- [ ] Empty favorite lists

### Integration Tests
- [ ] Play → Save → Play again workflow
- [ ] Search → Play → Save → Search again
- [ ] QuickPlay → Regular play → Compare
- [ ] Multiple lists in play screen

**Result:** ⬜ Pass ⬜ Fail  
**Notes:**

---

## Performance Tests

### Responsiveness
- [ ] Lists scroll smoothly
- [ ] Filter responds quickly
- [ ] No lag when typing
- [ ] Resize is smooth

### Memory
- [ ] No memory leaks on long sessions
- [ ] Multiple play/stop cycles stable
- [ ] Filter doesn't slow down

**Result:** ⬜ Pass ⬜ Fail  
**Notes:**

---

## Documentation Verification

- [ ] `BUG_FIXES_COMPLETE.md` is accurate
- [ ] `FIXES_SUMMARY.md` matches implementation
- [ ] `VISUAL_FIXES_GUIDE.md` examples work
- [ ] Code comments are clear
- [ ] User-facing messages are helpful

---

## Final Checklist

### Before Declaring Complete
- [ ] All 5 issues tested
- [ ] All tests passed
- [ ] No new issues discovered
- [ ] Documentation reviewed
- [ ] Ready for user testing

### If Issues Found
- [ ] Document issue clearly
- [ ] Determine if blocking or minor
- [ ] Create fix if needed
- [ ] Re-test affected areas

---

## Sign-Off

**Tester:** _______________  
**Date:** _______________  
**Terminal:** _______________ (OS and size)  
**MPV Version:** _______________

**Overall Result:** ⬜ PASS ⬜ FAIL

**Additional Notes:**

---

**Critical Issues (Must Fix Before Release):**
1. 
2.

**Minor Issues (Can Fix Later):**
1.
2.

**Suggestions for Improvement:**
1.
2.

---

## Quick Reference

**Build:**
```bash
./build_and_verify.sh
```

**Run:**
```bash
./tera
```

**Check Processes:**
```bash
ps aux | grep mpv
```

**Test Data:**
- Use "jazz" tag for searches (usually has good results)
- Test with empty and full favorite lists
- Try both short and long station names
