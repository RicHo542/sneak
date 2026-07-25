package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/richo542/sneak/internal/config"
)

type JiraProviderClient struct {
	cfg    *config.Provider
	client *http.Client
}

func (c *JiraProviderClient) TestConnection() error {
	host := strings.TrimRight(c.cfg.Host, "/")
	url := host + "/rest/api/2/myself"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.cfg.Username, c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	return TestRequest(c.client, req)
}
