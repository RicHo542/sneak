package azure

import (
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
