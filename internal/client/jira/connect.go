package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/richo542/sneak/internal/client/objects"
)

// jiraMyselfResponse is the relevant part of the /myself response.
type jiraMyselfResponse struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
}

func (c *JiraProviderClient) TestConnection() error {
	apiURL := c.Endpoints.testEndpoint()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.Cfg.Username, c.Cfg.Token)
	req.Header.Set("Accept", "application/json")

	return testRequest(c.Client, req)
}

// GetUserIdent returns the accountId of the authenticated user.
// The user id will later be used to assign items to the current user
func (c *JiraProviderClient) GetUserIdent() (*objects.UserInfo, error) {
	apiURL := c.Endpoints.myselfEndpoint()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.Cfg.Username, c.Cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var myself jiraMyselfResponse
	if err := json.Unmarshal(respBody, &myself); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if myself.AccountID == "" {
		return nil, fmt.Errorf("could not determine current user's accountId")
	}

	return &objects.UserInfo{
		UserHandle:  myself.AccountID,
		DisplayName: myself.DisplayName,
	}, nil
}

func testRequest(client *http.Client, req *http.Request) error {
	resp, err := client.Do(req)
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
