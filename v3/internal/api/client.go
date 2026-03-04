package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var baseURL = "https://de1.api.radio-browser.info/json/stations"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SearchByTag(ctx context.Context, tag string) ([]Station, error) {
	form := url.Values{}
	form.Add("tag", tag)

	return c.doSearch(ctx, form)
}

func (c *Client) doSearch(ctx context.Context, form url.Values) ([]Station, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		baseURL+"/search",
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<10))
		return nil, fmt.Errorf("search request failed: %s (body: %s)", resp.Status, string(body))
	}

	// Limit response to 10 MB to prevent memory exhaustion from a misbehaving server.
	var stations []Station
	if err := json.NewDecoder(io.LimitReader(resp.Body, 10<<20)).Decode(&stations); err != nil {
		return nil, err
	}

	return stations, nil
}

// VoteResult represents the response from voting for a station
type VoteResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// Vote increases the vote count for a station by one
// Note: Can only vote once per IP per station every 10 minutes
func (c *Client) Vote(ctx context.Context, stationUUID string) (*VoteResult, error) {
	// Strip /stations suffix and add /vote endpoint
	voteURL := strings.TrimSuffix(baseURL, "/stations") + "/vote/" + url.PathEscape(stationUUID)

	req, err := http.NewRequestWithContext(ctx, "POST", voteURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("vote request failed with status: %d", resp.StatusCode)
	}

	var result VoteResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
