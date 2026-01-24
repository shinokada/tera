You can access /Users/shinichiokada/Terminal-Tools/tera.
Please read CLAUDE.md first.

In the last session, we started coding for the Search Screen.
However there are quite few problems.
1. Test fails as the following:
```
➜  tera git:(golang) ✗ ./run_search_tests.sh
================================
TERA Search Screen Test Suite
================================

\033[1;33mRunning API Search Tests...\033[0m
----------------------------
# github.com/shinokada/tera/internal/api [github.com/shinokada/tera/internal/api.test]
internal/api/search_test.go:39:2: cannot assign to baseURL (neither addressable nor a map index expression)
internal/api/search_test.go:40:17: cannot assign to baseURL (neither addressable nor a map index expression)
internal/api/search_test.go:86:2: cannot assign to baseURL (neither addressable nor a map index expression)
internal/api/search_test.go:87:17: cannot assign to baseURL (neither addressable nor a map index expression)
internal/api/search_test.go:129:2: cannot assign to baseURL (neither addressable nor a map index expression)
internal/api/search_test.go:130:17: cannot assign to baseURL (neither addressable nor a map index expression)
internal/api/search_test.go:171:2: cannot assign to baseURL (neither addressable nor a map index expression)
internal/api/search_test.go:172:17: cannot assign to baseURL (neither addressable nor a map index expression)
internal/api/search_test.go:312:2: cannot assign to baseURL (neither addressable nor a map index expression)
internal/api/search_test.go:313:17: cannot assign to baseURL (neither addressable nor a map index expression)
internal/api/search_test.go:313:17: too many errors
FAIL    github.com/shinokada/tera/internal/api [build failed]
FAIL
✗ API Search Tests Failed

\033[1;33mRunning UI Search Tests...\033[0m
----------------------------
# github.com/shinokada/tera/internal/ui [github.com/shinokada/tera/internal/ui.test]
internal/ui/search_test.go:429:6: contains redeclared in this block
        internal/ui/play_station_test.go:272:6: other declaration of contains
internal/ui/search_test.go:435:6: findSubstring redeclared in this block
        internal/ui/play_station_test.go:277:6: other declaration of findSubstring
internal/ui/search.go:356:45: too many arguments in call to m.player.Play
        have (string, string)
        want (*api.Station)
internal/ui/search.go:361:12: m.player.Wait undefined (type *player.MPVPlayer has no field or method Wait)
internal/ui/search.go:450:17: undefined: subtleStyle
internal/ui/search.go:464:17: undefined: subtleStyle
internal/ui/search.go:477:18: undefined: subtleStyle
internal/ui/search.go:481:18: undefined: subtleStyle
internal/ui/search.go:494:17: undefined: subtleStyle
internal/ui/search.go:553:16: undefined: subtleStyle
internal/ui/search.go:553:16: too many errors
FAIL    github.com/shinokada/tera/internal/ui [build failed]
FAIL
✗ UI Search Tests Failed

================================
Test Summary
================================
Total Test Suites: 2
Passed: 0
Failed: 2

Some tests failed. Please review the output above.
```

2. Relating to the #1, golang/internal/api/search_test.go has many errors about baseURL on VSCode.
- annot assign to baseURL (neither addressable nor a map index expression)compilerUnassignableOperand
const baseURL untyped string = "https://de1.api.radio-browser.info/json/stations"

3. internal/ui/play_station_test.go has errors on VSCode.
- contains redeclared in this block (see details)compilerDuplicateDecl
search_test.go(429, 6):
func contains(s string, substr string) bool
Helper function
- findSubstring redeclared in this block (see details)compilerDuplicateDecl
search_test.go(435, 6):
func findSubstring(s string, substr string) bool

4. internal/ui/search_test.go has errors on VSCode.
- undefined: api.ErrEmptyResponsecompilerUndeclaredImportedName
- undefined: tea.ListItemcompilerUndeclaredImportedName
- contains redeclared in this blockcompilerDuplicateDecl
play_station_test.go(272, 6): other declaration of contains
- findSubstring redeclared in this blockcompilerDuplicateDecl
play_station_test.go(277, 6): other declaration of findSubstring

5. internal/ui/search.go has errors:
- too many arguments in call to m.player.Play
	have (string, string)
	want (*api.Station)compilerWrongArgCount
field Name string `json:"name"`
- m.player.Wait undefined (type *player.MPVPlayer has no field or method Wait)compilerMissingFieldOrMethod
- undefined: subtleStyle
- undefined: boldStyle