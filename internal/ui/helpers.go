package ui

import (
	"fmt"

	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/storage"
)

// formatSleepCountdown returns the decorated countdown string (e.g. "💤 Stops in 12:34")
// when countdown is non-empty, or an empty string when no timer is active.
func formatSleepCountdown(countdown string) string {
	if countdown == "" {
		return ""
	}
	return "💤 " + countdown
}

// renderNowPlayingBar returns a single-line now-playing banner for use at the
// bottom of any screen when ContinueOnNavigate is active and a station is
// playing at the app level. Returns an empty string when station is nil.
// Format:  ♫ Jazz FM  [context]  ·  Vol: 80%  ·  Space: Pause  //*: Vol  Esc: Stop
func renderNowPlayingBar(station *api.Station, contextLabel string, vol int) string {
	if station == nil {
		return ""
	}
	name := station.TrimName()
	bar := fmt.Sprintf("♫ %s", name)
	if contextLabel != "" {
		bar += fmt.Sprintf("  [%s]", contextLabel)
	}
	bar += fmt.Sprintf("  ·  Vol: %d%%  ·  Space: Pause  //*: Vol  Esc: Stop", vol)
	return successStyle().Render(bar)
}

// hydrateStations hydrates station metadata for a list of UUIDs from the cache.
// Stations with no cached entry get the UUID as a fallback name.
func hydrateStations(mm *storage.MetadataManager, uuids []string) []api.Station {
	stations := make([]api.Station, 0, len(uuids))
	for _, uuid := range uuids {
		var s api.Station
		s.StationUUID = uuid
		if mm != nil {
			if cached := mm.GetCachedStation(uuid); cached != nil {
				s.Name = cached.Name
				s.Country = cached.Country
				s.Codec = cached.Codec
				s.Bitrate = cached.Bitrate
				s.URLResolved = cached.URL
			}
		}
		if s.Name == "" {
			s.Name = uuid
		}
		stations = append(stations, s)
	}
	return stations
}


