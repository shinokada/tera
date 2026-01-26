package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/shinokada/tera/internal/api"
)

type Storage struct {
	favoritePath string
}

func NewStorage(favoritePath string) *Storage {
	return &Storage{favoritePath: favoritePath}
}

func (s *Storage) LoadList(ctx context.Context, name string) (*FavoritesList, error) {
	path := filepath.Join(s.favoritePath, name+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var stations []api.Station
	if err := json.Unmarshal(data, &stations); err != nil {
		return nil, err
	}

	return &FavoritesList{
		Name:     name,
		Stations: stations,
	}, nil
}

// SaveList saves a favorites list to disk
func (s *Storage) SaveList(ctx context.Context, list *FavoritesList) error {
	path := filepath.Join(s.favoritePath, list.Name+".json")

	data, err := json.MarshalIndent(list.Stations, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// AddStation adds a station to a list, checking for duplicates by UUID
func (s *Storage) AddStation(ctx context.Context, listName string, station api.Station) error {
	// Load existing list
	list, err := s.LoadList(ctx, listName)
	if err != nil {
		// If list doesn't exist, create new one
		if os.IsNotExist(err) {
			list = &FavoritesList{
				Name:     listName,
				Stations: []api.Station{},
			}
		} else {
			return err
		}
	}

	// Check for duplicates
	for _, existing := range list.Stations {
		if existing.StationUUID == station.StationUUID {
			return ErrDuplicateStation
		}
	}

	// Add station
	list.Stations = append(list.Stations, station)

	// Save
	return s.SaveList(ctx, list)
}

// StationExists checks if a station exists in a list
func (s *Storage) StationExists(ctx context.Context, listName string, stationUUID string) (bool, error) {
	list, err := s.LoadList(ctx, listName)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	for _, station := range list.Stations {
		if station.StationUUID == stationUUID {
			return true, nil
		}
	}

	return false, nil
}

// RemoveStation removes a station from a list by UUID
func (s *Storage) RemoveStation(ctx context.Context, listName string, stationUUID string) error {
	// Load existing list
	list, err := s.LoadList(ctx, listName)
	if err != nil {
		return err
	}

	// Find and remove the station
	found := false
	newStations := make([]api.Station, 0, len(list.Stations))
	for _, station := range list.Stations {
		if station.StationUUID != stationUUID {
			newStations = append(newStations, station)
		} else {
			found = true
		}
	}

	if !found {
		return ErrStationNotFound
	}

	// Update list
	list.Stations = newStations

	// Save
	return s.SaveList(ctx, list)
}
