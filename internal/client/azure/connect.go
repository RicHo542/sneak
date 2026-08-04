package azure

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *AzureProviderClient) TestConnection() error {
	apiURL := c.Endpoints.testEndpoint()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth("", c.Cfg.Token)
	req.Header.Set("Accept", "application/json")

	return c.testRequest(req)
}

func (c *AzureProviderClient) testRequest(req *http.Request) error {
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	io.Copy(io.Discard, resp.Body)
	return nil
}
