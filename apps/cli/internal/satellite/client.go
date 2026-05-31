package satellite

import (
	"fmt"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

const DefaultBaseURL = "https://celestrak.org/NORAD/elements/gp.php"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) buildURL(queryKey string, queryValue string) (string, error) {
	baseURL, err := url.Parse(c.baseURL)	
	if err != nil {
		return "", err
	}

	query := baseURL.Query()
	query.Set(queryKey, queryValue)
	query.Set("FORMAT", "JSON")

	baseURL.RawQuery = query.Encode()
	
	return baseURL.String(), nil
}

func (c *Client) Fetch(ctx context.Context, queryKey string, queryValue string) ([]GPElement, error) {
	apiURL, err := c.buildURL(queryKey, queryValue)
	if err != nil {
		return nil, fmt.Errorf("build CelesTrak URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create CelesTrak request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call CelesTrak API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CelesTrak returned status %s", resp.Status)
	}

	var elements []GPElement
	if err := json.NewDecoder(resp.Body).Decode(&elements); err != nil {
		return nil, fmt.Errorf("decode CelesTrak response: %w", err)
	}

	return elements, nil
}
