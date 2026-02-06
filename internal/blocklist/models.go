package blocklist

import (
	"errors"
	"time"
)

// Errors
var (
	ErrStationAlreadyBlocked = errors.New("station is already blocked")
	ErrStationNotBlocked     = errors.New("station is not blocked")
)

// BlockedStation represents a blocked radio station with metadata
type BlockedStation struct {
	StationUUID string    `json:"stationuuid"`
	Name        string    `json:"name"`
	Tags        string    `json:"tags,omitempty"`
	Country     string    `json:"country,omitempty"`
	CountryCode string    `json:"countrycode,omitempty"`
	State       string    `json:"state,omitempty"`
	Language    string    `json:"language,omitempty"`
	Codec       string    `json:"codec,omitempty"`
	Bitrate     int       `json:"bitrate,omitempty"`
	BlockedAt   time.Time `json:"blocked_at"`
}

// Blocklist represents the complete blocklist data structure
type Blocklist struct {
	Version         string           `json:"version"`
	BlockedStations []BlockedStation `json:"blocked_stations"`
	BlockRules      []BlockRule      `json:"block_rules,omitempty"`
}

// BlockWarningThresholds defines when to show warnings
const (
	BlockWarningThreshold = 100
	BlockLargeThreshold   = 500
)
