package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
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
	defer resp.Body.Close()

	var stations []Station
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return nil, err
	}

	return stations, nil
}
