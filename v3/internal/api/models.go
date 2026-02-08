package api

import "strings"

// Station represents a radio station from Radio Browser API
type Station struct {
	StationUUID string `json:"stationuuid"`
	Name        string `json:"name"`
	URLResolved string `json:"url_resolved"`
	Tags        string `json:"tags"`
	Country     string `json:"country"`
	CountryCode string `json:"countrycode"`
	State       string `json:"state"`
	Language    string `json:"language"`
	Votes       int    `json:"votes"`
	Codec       string `json:"codec"`
	Bitrate     int    `json:"bitrate"`
	Volume      *int   `json:"volume,omitempty"` // Per-station volume (0-100), nil means use default
}

// TrimName returns station name with whitespace trimmed
func (s *Station) TrimName() string {
	return strings.TrimSpace(s.Name)
}

// SetVolume sets the station's volume (0-100)
func (s *Station) SetVolume(vol int) {
	s.Volume = &vol
}

// GetVolume returns the station's volume or -1 if not set
func (s *Station) GetVolume() int {
	if s.Volume == nil {
		return -1
	}
	return *s.Volume
}
