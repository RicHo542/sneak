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
	cfg *config.Provider, client *http.Client, lctx *config.LocalContext,
) *AzureProviderClient {
	host := strings.TrimRight(cfg.Host, "/")

	// In case nothing was initialized locally in .sneak,
	// the provider client will not have the Remote.Project information yet.
	// Endpoints requiring it will not be used for testing the connection though
	project := ""
	if lctx != nil && lctx.Remote.Project != "" {
		project = lctx.Remote.Project
	}

	return &AzureProviderClient{
		Cfg:    cfg,
		Client: client,
		Endpoints: &AzureEndpoints{
			host:         host,
			apiVersion:   "7.0",
			organization: cfg.Organization,
			project:      project,
		},
	}
}

func azureWorkItemToObject(wi azureBatchWorkItem) objects.WorkItem {
	item := objects.WorkItem{
		ID:      strconv.Itoa(wi.ID),
		Key:     fmt.Sprintf("%d", wi.ID),
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
	req.SetBasicAuth(
		"",
		c.Cfg.Token,
	)
}
