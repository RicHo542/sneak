package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type azureIdentityValue struct {
	Value string `json:"$value"`
}

type azureIdentityProperties struct {
	Account *azureIdentityValue `json:"Account"`
	Mail    *azureIdentityValue `json:"Mail"`
}

type azureAuthenticatedUser struct {
	DisplayName string                  `json:"displayName"`
	Properties  azureIdentityProperties `json:"properties"`
}

type azureConnectionData struct {
	AuthenticatedUser azureAuthenticatedUser `json:"authenticatedUser"`
}

func (c *AzureProviderClient) TestConnection() error {
	apiURL := c.Endpoints.testEndpoint()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	c.SetAuthHeader(req)
	req.Header.Set("Accept", "application/json")

	return c.testRequest(req)
}

// GetUserIdent returns the account id, the user mail or display name of the
// authenticated user to be used for work item assignments.
func (c *AzureProviderClient) GetUserIdent() (string, error) {
	apiURL := c.Endpoints.connectionDataEndpoint()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("bad request: %w", err)
	}
	c.SetAuthHeader(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var data azureConnectionData
	if err := json.Unmarshal(respBody, &data); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	user := data.AuthenticatedUser

	// Prioritize AccountIds if available. Email second and displayed username last
	switch {
	case user.Properties.Account != nil && user.Properties.Account.Value != "":
		return user.Properties.Account.Value, nil
	case user.Properties.Mail != nil && user.Properties.Mail.Value != "":
		return user.Properties.Mail.Value, nil
	case user.DisplayName != "":
		return user.DisplayName, nil
	}

	return "", fmt.Errorf("could not determine current user's identity handle")
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
