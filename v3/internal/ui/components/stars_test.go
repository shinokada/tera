package components

import (
	"testing"
)

func TestStarRenderer(t *testing.T) {
	t.Run("NewStarRenderer_Unicode", func(t *testing.T) {
		sr := NewStarRenderer(true)
		if sr.filledStar != "★" {
			t.Errorf("Expected unicode filled star ★, got %s", sr.filledStar)
		}
		if sr.emptyStar != "☆" {
			t.Errorf("Expected unicode empty star ☆, got %s", sr.emptyStar)
		}
	})

	t.Run("NewStarRenderer_ASCII", func(t *testing.T) {
		sr := NewStarRenderer(false)
		if sr.filledStar != "*" {
			t.Errorf("Expected ASCII filled star *, got %s", sr.filledStar)
		}
		if sr.emptyStar != "-" {
			t.Errorf("Expected ASCII empty star -, got %s", sr.emptyStar)
		}
	})

	t.Run("RenderPlain", func(t *testing.T) {
		sr := NewStarRenderer(true)

		tests := []struct {
			rating   int
			expected string
		}{
			{0, "☆ ☆ ☆ ☆ ☆"},
			{1, "★ ☆ ☆ ☆ ☆"},
			{2, "★ ★ ☆ ☆ ☆"},
			{3, "★ ★ ★ ☆ ☆"},
			{4, "★ ★ ★ ★ ☆"},
			{5, "★ ★ ★ ★ ★"},
			{-1, "☆ ☆ ☆ ☆ ☆"},
			{6, "★ ★ ★ ★ ★"},
		}

		for _, tt := range tests {
			result := sr.RenderPlain(tt.rating)
			if result != tt.expected {
				t.Errorf("RenderPlain(%d) = %q, want %q", tt.rating, result, tt.expected)
			}
		}
	})

	t.Run("RenderPlain_ASCII", func(t *testing.T) {
		sr := NewStarRenderer(false)

		tests := []struct {
			rating   int
			expected string
		}{
			{0, "- - - - -"},
			{3, "* * * - -"},
			{5, "* * * * *"},
		}

		for _, tt := range tests {
			result := sr.RenderPlain(tt.rating)
			if result != tt.expected {
				t.Errorf("RenderPlain(%d) = %q, want %q", tt.rating, result, tt.expected)
			}
		}
	})

	t.Run("RenderCompactPlain", func(t *testing.T) {
		sr := NewStarRenderer(true)

		tests := []struct {
			rating   int
			expected string
		}{
			{0, ""},
			{1, "★"},
			{2, "★ ★"},
			{3, "★ ★ ★"},
			{4, "★ ★ ★ ★"},
			{5, "★ ★ ★ ★ ★"},
			{-1, ""},
			{6, ""},
		}

		for _, tt := range tests {
			result := sr.RenderCompactPlain(tt.rating)
			if result != tt.expected {
				t.Errorf("RenderCompactPlain(%d) = %q, want %q", tt.rating, result, tt.expected)
			}
		}
	})

	t.Run("Width", func(t *testing.T) {
		sr := NewStarRenderer(true)
		if sr.Width() != 9 {
			t.Errorf("Expected width 9, got %d", sr.Width())
		}
	})

	t.Run("FilledStar_EmptyStar", func(t *testing.T) {
		sr := NewStarRenderer(true)
		if sr.FilledStar() != "★" {
			t.Errorf("Expected ★, got %s", sr.FilledStar())
		}
		if sr.EmptyStar() != "☆" {
			t.Errorf("Expected ☆, got %s", sr.EmptyStar())
		}
	})

	t.Run("DefaultStarRenderer", func(t *testing.T) {
		sr := DefaultStarRenderer()
		if sr == nil {
			t.Fatal("Expected non-nil default renderer")
		}
		if sr.filledStar != "★" {
			t.Error("Default renderer should use unicode")
		}
	})
}
