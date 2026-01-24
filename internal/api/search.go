package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// SearchType represents the type of search query
type SearchType int

const (
	SearchByTag SearchType = iota
	SearchByName
	SearchByLanguage
	SearchByCountry
	SearchByState
	SearchAdvanced
)

// SearchParams holds parameters for search requests
type SearchParams struct {
	Tag      string
	Name     string
	Language string
	Country  string
	State    string
	// Advanced search can combine multiple parameters
	TagExact     bool
	NameExact    bool
	Order        string // votes, clickcount, bitrate, name
	Reverse      bool
	Limit        int
	Offset       int
	HideBroken   bool
}

// Search performs a search based on the given parameters
func (c *Client) Search(ctx context.Context, params SearchParams) ([]Station, error) {
	form := c.buildFormValues(params)
	return c.doSearch(ctx, form)
}

// buildFormValues constructs form values from search parameters
func (c *Client) buildFormValues(params SearchParams) url.Values {
	form := url.Values{}

	// Add search parameters
	if params.Tag != "" {
		form.Add("tag", strings.TrimSpace(params.Tag))
		if params.TagExact {
			form.Add("tagExact", "true")
		}
	}
	if params.Name != "" {
		form.Add("name", strings.TrimSpace(params.Name))
		if params.NameExact {
			form.Add("nameExact", "true")
		}
	}
	if params.Language != "" {
		form.Add("language", strings.TrimSpace(params.Language))
	}
	if params.Country != "" {
		form.Add("country", strings.TrimSpace(params.Country))
	}
	if params.State != "" {
		form.Add("state", strings.TrimSpace(params.State))
	}

	// Add ordering
	if params.Order != "" {
		form.Add("order", params.Order)
	} else {
		form.Add("order", "votes")
	}

	if params.Reverse {
		form.Add("reverse", "true")
	}

	// Add limit and offset
	if params.Limit > 0 {
		form.Add("limit", fmt.Sprintf("%d", params.Limit))
	} else {
		form.Add("limit", "100")
	}

	if params.Offset > 0 {
		form.Add("offset", fmt.Sprintf("%d", params.Offset))
	}

	// Hide broken stations by default
	form.Add("hidebroken", "true")

	return form
}

// SearchByName searches for stations by name
func (c *Client) SearchByName(ctx context.Context, name string) ([]Station, error) {
	params := SearchParams{
		Name:       strings.TrimSpace(name),
		Order:      "votes",
		Reverse:    true,
		Limit:      100,
		HideBroken: true,
	}
	return c.Search(ctx, params)
}

// SearchByLanguage searches for stations by language
func (c *Client) SearchByLanguage(ctx context.Context, language string) ([]Station, error) {
	params := SearchParams{
		Language:   strings.TrimSpace(language),
		Order:      "votes",
		Reverse:    true,
		Limit:      100,
		HideBroken: true,
	}
	return c.Search(ctx, params)
}

// SearchByCountry searches for stations by country code
func (c *Client) SearchByCountry(ctx context.Context, country string) ([]Station, error) {
	params := SearchParams{
		Country:    strings.TrimSpace(country),
		Order:      "votes",
		Reverse:    true,
		Limit:      100,
		HideBroken: true,
	}
	return c.Search(ctx, params)
}

// SearchByState searches for stations by state
func (c *Client) SearchByState(ctx context.Context, state string) ([]Station, error) {
	params := SearchParams{
		State:      strings.TrimSpace(state),
		Order:      "votes",
		Reverse:    true,
		Limit:      100,
		HideBroken: true,
	}
	return c.Search(ctx, params)
}

// SearchAdvanced performs an advanced search with multiple criteria
func (c *Client) SearchAdvanced(ctx context.Context, params SearchParams) ([]Station, error) {
	// Ensure defaults for advanced search
	if params.Order == "" {
		params.Order = "votes"
	}
	if params.Limit == 0 {
		params.Limit = 100
	}
	params.HideBroken = true

	return c.Search(ctx, params)
}

// doSearchWithEndpoint performs a search using a specific endpoint (for special search types)
func (c *Client) doSearchWithEndpoint(ctx context.Context, endpoint string, form url.Values) ([]Station, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		endpoint,
		bytes.NewBufferString(form.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var stations []Station
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return nil, err
	}

	return stations, nil
}
