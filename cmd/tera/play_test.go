package main

import (
	"testing"
	"time"
)

// -----------------------------------------------------------------
// parseFavArgs
// -----------------------------------------------------------------

func TestParseFavArgs_Defaults(t *testing.T) {
	listName, n := parseFavArgs([]string{})
	if listName != "My-favorites" {
		t.Errorf("expected My-favorites, got %q", listName)
	}
	if n != 1 {
		t.Errorf("expected n=1, got %d", n)
	}
}

func TestParseFavArgs_ListOnly(t *testing.T) {
	listName, n := parseFavArgs([]string{"jazz"})
	if listName != "jazz" {
		t.Errorf("expected jazz, got %q", listName)
	}
	if n != 1 {
		t.Errorf("expected n=1, got %d", n)
	}
}

func TestParseFavArgs_ListAndN(t *testing.T) {
	listName, n := parseFavArgs([]string{"jazz", "3"})
	if listName != "jazz" {
		t.Errorf("expected jazz, got %q", listName)
	}
	if n != 3 {
		t.Errorf("expected n=3, got %d", n)
	}
}

func TestParseFavArgs_NumericOnlyTreatedAsN(t *testing.T) {
	// A bare integer with no list name should use My-favorites and set n
	listName, n := parseFavArgs([]string{"5"})
	if listName != "My-favorites" {
		t.Errorf("expected My-favorites, got %q", listName)
	}
	if n != 5 {
		t.Errorf("expected n=5, got %d", n)
	}
}

func TestParseFavArgs_ListWithNonNumericN(t *testing.T) {
	// If second arg is not a number, n defaults to 1
	listName, n := parseFavArgs([]string{"jazz", "abc"})
	if listName != "jazz" {
		t.Errorf("expected jazz, got %q", listName)
	}
	if n != 1 {
		t.Errorf("expected n=1, got %d", n)
	}
}

// -----------------------------------------------------------------
// parseNArg
// -----------------------------------------------------------------

func TestParseNArg_Default(t *testing.T) {
	n := parseNArg([]string{})
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestParseNArg_Explicit(t *testing.T) {
	n := parseNArg([]string{"4"})
	if n != 4 {
		t.Errorf("expected 4, got %d", n)
	}
}

func TestParseNArg_InvalidFallsBackToOne(t *testing.T) {
	n := parseNArg([]string{"abc"})
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestParseNArg_ZeroFallsBackToOne(t *testing.T) {
	// Zero is not a valid 1-based index
	n := parseNArg([]string{"0"})
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestParseNArg_NegativeFallsBackToOne(t *testing.T) {
	n := parseNArg([]string{"-3"})
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

// -----------------------------------------------------------------
// duration parsing (via time.ParseDuration — exercised directly)
// -----------------------------------------------------------------

func TestParseDuration_Valid(t *testing.T) {
	cases := []struct {
		input    string
		expected time.Duration
	}{
		{"30s", 30 * time.Second},
		{"10m", 10 * time.Minute},
		{"1h", time.Hour},
		{"1h30m", 90 * time.Minute},
	}
	for _, tc := range cases {
		d, err := time.ParseDuration(tc.input)
		if err != nil {
			t.Errorf("unexpected error for %q: %v", tc.input, err)
		}
		if d != tc.expected {
			t.Errorf("for %q: expected %v, got %v", tc.input, tc.expected, d)
		}
	}
}

func TestParseDuration_Invalid(t *testing.T) {
	invalids := []string{"2x", "abc", "10", "-5m"}
	for _, s := range invalids {
		d, err := time.ParseDuration(s)
		if err == nil && d > 0 {
			t.Errorf("expected error or non-positive duration for %q, got %v", s, d)
		}
	}
}

// -----------------------------------------------------------------
// lucky keyword joining
// -----------------------------------------------------------------

func TestLuckyKeyword_SingleWord(t *testing.T) {
	args := []string{"lucky", "jazz"}
	keyword := joinLuckyKeyword(args[1:])
	if keyword != "jazz" {
		t.Errorf("expected jazz, got %q", keyword)
	}
}

func TestLuckyKeyword_MultiWord(t *testing.T) {
	args := []string{"lucky", "smooth", "jazz"}
	keyword := joinLuckyKeyword(args[1:])
	if keyword != "smooth jazz" {
		t.Errorf("expected 'smooth jazz', got %q", keyword)
	}
}

func TestLuckyKeyword_ThreeWords(t *testing.T) {
	args := []string{"lucky", "80s", "classic", "rock"}
	keyword := joinLuckyKeyword(args[1:])
	if keyword != "80s classic rock" {
		t.Errorf("expected '80s classic rock', got %q", keyword)
	}
}

// -----------------------------------------------------------------
// truncate
// -----------------------------------------------------------------

func TestTruncate_ShortString(t *testing.T) {
	result := truncate("Jazz FM", 40)
	if result != "Jazz FM" {
		t.Errorf("expected 'Jazz FM', got %q", result)
	}
}

func TestTruncate_ExactLength(t *testing.T) {
	s := "12345678901234567890123456789012345678901" // 41 chars
	result := truncate(s, 40)
	runes := []rune(result)
	if len(runes) > 40 {
		t.Errorf("expected at most 40 runes, got %d", len(runes))
	}
	if runes[len(runes)-1] != '…' {
		t.Errorf("expected trailing ellipsis, got %q", string(runes[len(runes)-1]))
	}
}

func TestTruncate_Unicode(t *testing.T) {
	// Each character here is a multi-byte rune
	s := "日本語ラジオ放送局テスト名前" // 13 runes
	result := truncate(s, 10)
	runes := []rune(result)
	if len(runes) > 10 {
		t.Errorf("expected at most 10 runes, got %d", len(runes))
	}
}
