package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/richo542/sneak/internal/config"
)

type AzureProviderClient struct {
	cfg    *config.Provider
	client *http.Client
}

func (c *AzureProviderClient) TestConnection() error {
	host := strings.TrimRight(c.cfg.Host, "/")
	url := host + "/_apis/profile/profiles/me?api-version=7.0"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth("", c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	return TestRequest(c.client, req)
}
