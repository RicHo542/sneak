package azure

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/richo542/sneak/internal/client/objects"
	"github.com/richo542/sneak/internal/config"
)

type AzureProviderClient struct {
	Cfg       *config.Provider
	Client    *http.Client
	Endpoints *AzureEndpoints
}

func NewAzureProviderClient(
	cfg *config.Provider, client *http.Client,
) *AzureProviderClient {
	host := strings.TrimRight(cfg.Host, "/")

	return &AzureProviderClient{
		Cfg:       cfg,
		Client:    client,
		Endpoints: &AzureEndpoints{host: host, apiVersion: "7.0"},
	}
}

func azureWorkItemToObject(wi azureBatchWorkItem) objects.WorkItem {
	item := objects.WorkItem{
		ID:      strconv.Itoa(wi.ID),
		Key:     fmt.Sprintf("#%d", wi.ID),
		Summary: wi.Fields.Title,
		Status:  wi.Fields.State,
		Type:    wi.Fields.Type,
	}
	if wi.Fields.Assignee != nil {
		item.Assignee = wi.Fields.Assignee.DisplayName
	}
	return item
}

func (c *AzureProviderClient) SetAuthHeader(req *http.Request) {
	tokenStr := fmt.Sprintf(":%s", c.Cfg.Token)
	req.SetBasicAuth(
		"",
		base64.StdEncoding.EncodeToString([]byte(tokenStr)),
	)
}
