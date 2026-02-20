package components

import (
	"strings"
	"testing"
)

func TestRenderPillContainsTag(t *testing.T) {
	r := NewTagRenderer()
	pill := r.RenderPill("chill vibes")
	if !strings.Contains(pill, "chill vibes") {
		t.Errorf("RenderPill: expected pill to contain tag text, got %q", pill)
	}
}

func TestRenderPillsEmpty(t *testing.T) {
	r := NewTagRenderer()
	if got := r.RenderPills(nil); got != "" {
		t.Errorf("RenderPills(nil): expected empty string, got %q", got)
	}
	if got := r.RenderPills([]string{}); got != "" {
		t.Errorf("RenderPills([]): expected empty string, got %q", got)
	}
}

func TestRenderPillsSingleTag(t *testing.T) {
	r := NewTagRenderer()
	got := r.RenderPills([]string{"focus"})
	if !strings.Contains(got, "focus") {
		t.Errorf("RenderPills single: expected 'focus' in output, got %q", got)
	}
}

func TestRenderPillsOverflow(t *testing.T) {
	r := NewTagRenderer()
	tags := []string{"alpha", "beta", "gamma", "delta"} // 4 > maxDisplay(3)
	got := r.RenderPills(tags)
	if !strings.Contains(got, "alpha") || !strings.Contains(got, "beta") || !strings.Contains(got, "gamma") {
		t.Errorf("RenderPills overflow: first 3 tags missing from %q", got)
	}
	if strings.Contains(got, "delta") {
		t.Errorf("RenderPills overflow: 4th tag 'delta' should not appear in %q", got)
	}
	if !strings.Contains(got, "+1") {
		t.Errorf("RenderPills overflow: expected '+1' overflow indicator in %q", got)
	}
}

func TestRenderPillsNoOverflowAtExactMax(t *testing.T) {
	r := NewTagRenderer()
	tags := []string{"one", "two", "three"} // exactly maxDisplay
	got := r.RenderPills(tags)
	if strings.Contains(got, "+") {
		t.Errorf("RenderPills exact max: should not show overflow indicator, got %q", got)
	}
	for _, tag := range tags {
		if !strings.Contains(got, tag) {
			t.Errorf("RenderPills exact max: missing tag %q in output %q", tag, got)
		}
	}
}

func TestRenderListEmpty(t *testing.T) {
	r := NewTagRenderer()
	got := r.RenderList(nil)
	// Should return a "No tags" placeholder, not empty
	if got == "" {
		t.Error("RenderList(nil): expected non-empty 'No tags' string")
	}
}

func TestRenderListShowsAllTags(t *testing.T) {
	r := NewTagRenderer()
	tags := []string{"jazz", "late night", "coding"}
	got := r.RenderList(tags)
	for _, tag := range tags {
		if !strings.Contains(got, tag) {
			t.Errorf("RenderList: missing tag %q in output %q", tag, got)
		}
	}
}

func TestRenderPillsLargeOverflow(t *testing.T) {
	r := NewTagRenderer()
	tags := []string{"a", "b", "c", "d", "e", "f", "g"} // 7 tags
	got := r.RenderPills(tags)
	if !strings.Contains(got, "+4") {
		t.Errorf("RenderPills large overflow: expected '+4', got %q", got)
	}
}
