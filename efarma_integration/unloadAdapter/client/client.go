package client

import (
	"efarma_integration/config"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	unloadUrl  string
	apiKey     string
	httpClient *http.Client
}

func New(cnf *config.HttpClientConfig) *Client {
	return &Client{
		unloadUrl: cnf.UnloadUrl,
		apiKey:    cnf.ApiKey,
		httpClient: &http.Client{
			Timeout: time.Duration(cnf.Timeout) * time.Second,
		},
	}
}

func (c *Client) Unload(data io.Reader) error {
	req, err := http.NewRequest(http.MethodPost, c.unloadUrl, data)
	if err != nil {
		return fmt.Errorf("create upload request failed: %v", err)
	}

	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("X-Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: %v", http.StatusText(resp.StatusCode))
	}
	return nil
}
