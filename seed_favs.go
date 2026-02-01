package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shinokada/tera/internal/api"
	"github.com/shinokada/tera/internal/storage"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	favPath := filepath.Join(home, ".tera", "favorites")
	if err := os.MkdirAll(favPath, 0700); err != nil {
		panic(err)
	}

	store := storage.NewStorage(favPath)
	ctx := context.Background()

	// Create My-favorites.json
	list := &storage.FavoritesList{
		Name: "My-favorites",
		Stations: []api.Station{
			{
				StationUUID: "960de132-0601-11e8-ae97-52543be04c81",
				Name:        "Radio Paradise Main Mix",
				URLResolved: "http://stream.radioparadise.com/aac-128",
				Votes:       15000,
				Country:     "The United States Of America",
				Codec:       "AAC",
				Bitrate:     128,
				Tags:        "eclectic,paradise,rock",
			},
		},
	}

	if err := store.SaveList(ctx, list); err != nil {
		fmt.Printf("Error saving list: %v\n", err)
	} else {
		fmt.Println("Created My-favorites.json")
	}
}
