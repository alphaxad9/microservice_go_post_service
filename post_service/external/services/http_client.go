package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HTTPClient struct {
	client *http.Client
	apiKey string
}

func NewHTTPClient(apiKey string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiKey: apiKey,
	}
}

func (h *HTTPClient) Get(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-Internal-Key", h.apiKey)
	req.Header.Set("User-Agent", "post-service/1.0")

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
