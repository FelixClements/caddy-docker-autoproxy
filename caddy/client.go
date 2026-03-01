package caddy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Client wraps the Caddy Admin API client
type Client struct {
	adminURL  string
	httpClient *http.Client
}

// NewClient creates a new Caddy API client using CADDY_URL env var or default
func NewClient() *Client {
	adminURL := os.Getenv("CADDY_URL")
	if adminURL == "" {
		adminURL = "http://localhost:2019"
	}
	return NewClientWithURL(adminURL)
}

// NewClientWithURL creates a new Caddy API client with a specific URL
func NewClientWithURL(adminURL string) *Client {
	return &Client{
		adminURL: adminURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// PushConfig pushes configuration to Caddy Admin API
func (c *Client) PushConfig(ctx context.Context, config map[string]interface{}) error {
	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	url := c.adminURL + "/load"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to push config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("caddy API returned status %d", resp.StatusCode)
	}

	return nil
}

// PushConfigToPath pushes config to a specific path (for incremental updates)
func (c *Client) PushConfigToPath(ctx context.Context, path string, config interface{}) error {
	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	url := c.adminURL + path

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to push config to %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("caddy API returned status %d for path %s", resp.StatusCode, path)
	}

	return nil
}

// GetConfig retrieves the current Caddy configuration
func (c *Client) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	url := c.adminURL + "/config"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("caddy API returned status %d", resp.StatusCode)
	}

	var config map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return config, nil
}
