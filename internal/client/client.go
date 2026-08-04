package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/richo542/sneak/internal/client/azure"
	"github.com/richo542/sneak/internal/client/jira"
	"github.com/richo542/sneak/internal/client/objects"
	"github.com/richo542/sneak/internal/config"
)

type ProviderClient interface {
	TestConnection() error
	ListWorkItems(ctx *config.Context, opts objects.ListOptions) ([]objects.WorkItem, error)
	DiscoverWorkflow(ctx *config.Context) (map[string]config.WorkflowMap, error)
	TransitionWorkItems(ctx *config.Context, items []*config.CacheItem, ref config.TransitionRef) error
	AddCommentToWorkItems(*config.Context, []*config.CacheItem, string) error
}

func NewProviderClient(p *config.Provider) (ProviderClient, error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	switch p.Type {
	case "jira":
		return jira.NewJiraProviderClient(
			p, &client,
		), nil
	case "azure":
		return azure.NewAzureProviderClient(
			p, &client,
		), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", p.Type)
	}
}
