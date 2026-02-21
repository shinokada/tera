package ui

import (
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/storage"
)

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


