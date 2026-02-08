package api

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		// Basic comparisons
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"v1 less than v2", "1.0.0", "1.0.1", -1},
		{"v1 greater than v2", "1.0.1", "1.0.0", 1},
		{"major version diff", "1.0.0", "2.0.0", -1},
		{"minor version diff", "1.1.0", "1.2.0", -1},

		// With v prefix
		{"v prefix on v1", "v1.0.0", "1.0.0", 0},
		{"v prefix on v2", "1.0.0", "v1.0.0", 0},
		{"v prefix on both", "v1.0.0", "v1.0.1", -1},

		// Dev version
		{"dev is older", "dev", "1.0.0", -1},
		{"any version newer than dev", "1.0.0", "dev", 1},
		{"dev vs dev", "dev", "dev", -1}, // dev is always treated as older

		// Pre-release versions
		{"rc less than release", "1.0.0-rc.1", "1.0.0", -1},
		{"release greater than rc", "1.0.0", "1.0.0-rc.1", 1},
		{"equal base with rc", "1.0.0-rc.1", "1.0.0-rc.2", 0}, // pre-release suffix not deeply compared

		// Different lengths
		{"short vs long", "1.0", "1.0.0", 0},
		{"long vs short", "1.0.0", "1.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		name        string
		current     string
		latest      string
		expectNewer bool
	}{
		{"newer available", "1.0.0", "1.1.0", true},
		{"up to date", "1.1.0", "1.1.0", false},
		{"ahead of latest", "1.2.0", "1.1.0", false},
		{"dev to release", "dev", "1.0.0", true},
		{"rc to release", "1.0.0-rc.1", "1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNewerVersion(tt.current, tt.latest)
			if result != tt.expectNewer {
				t.Errorf("IsNewerVersion(%q, %q) = %v, want %v", tt.current, tt.latest, result, tt.expectNewer)
			}
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.0.0", "1.0.0"},
		{"1.0.0", "1.0.0"},
		{"  v1.0.0  ", "1.0.0"},
		{"V1.0.0", "V1.0.0"}, // Only lowercase v is stripped
		{"dev", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeVersion(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"1.0.0", []int{1, 0, 0}},
		{"1.2.3", []int{1, 2, 3}},
		{"1.0.0-rc.1", []int{1, 0, 0}}, // Pre-release suffix ignored
		{"10.20.30", []int{10, 20, 30}},
		{"1", []int{1}},
		{"1.2", []int{1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseVersion(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseVersion(%q) length = %d, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("parseVersion(%q)[%d] = %d, want %d", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}
