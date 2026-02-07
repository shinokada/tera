package blocklist

import (
	"strings"
	
	"github.com/shinokada/tera/internal/api"
)

// BlockRule represents a rule for blocking stations
type BlockRule struct {
	Type  BlockRuleType `json:"type"`  // "country", "language", "tag"
	Value string        `json:"value"` // The value to match (e.g., "US", "arabic", "sports")
}

// BlockRuleType represents the type of blocking rule
type BlockRuleType string

const (
	BlockRuleCountry  BlockRuleType = "country"
	BlockRuleLanguage BlockRuleType = "language"
	BlockRuleTag      BlockRuleType = "tag"
)

// BlockRules represents a collection of blocking rules
type BlockRules struct {
	Rules []BlockRule `json:"rules"`
}

// Matches checks if a station matches this rule
func (r BlockRule) Matches(station *api.Station) bool {
	if station == nil {
		return false
	}

	switch r.Type {
	case BlockRuleCountry:
		// Match against both Country and CountryCode
		return strings.EqualFold(station.Country, r.Value) ||
			strings.EqualFold(station.CountryCode, r.Value)

	case BlockRuleLanguage:
		// Match if language is present in comma-separated languages (case-insensitive)
		languages := strings.Split(station.Language, ",")
		for _, lang := range languages {
			if strings.EqualFold(strings.TrimSpace(lang), r.Value) {
				return true
			}
		}
		return false

	case BlockRuleTag:
		// Match if tag is present in comma-separated tags (case-insensitive)
		tags := strings.Split(station.Tags, ",")
		for _, tag := range tags {
			if strings.EqualFold(strings.TrimSpace(tag), r.Value) {
				return true
			}
		}
		return false

	default:
		return false
	}
}

// MatchesAny checks if a station matches any rule in the collection
func (rules BlockRules) MatchesAny(station *api.Station) bool {
	for _, rule := range rules.Rules {
		if rule.Matches(station) {
			return true
		}
	}
	return false
}

// String returns a human-readable representation of the rule
func (r BlockRule) String() string {
	switch r.Type {
	case BlockRuleCountry:
		return "Country: " + r.Value
	case BlockRuleLanguage:
		return "Language: " + r.Value
	case BlockRuleTag:
		return "Tag: " + r.Value
	default:
		return "Unknown rule"
	}
}
