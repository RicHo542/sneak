package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/richo542/sneak/internal/config"
)

type WorkItem struct {
	ID       string
	Key      string
	Summary  string
	Status   string
	Type     string
	Assignee string
}

type ListOptions struct {
	Bindings []string
	Types    []string
}

type ProviderClient interface {
	TestConnection() error
	ListWorkItems(ctx *config.Context, opts ListOptions) ([]WorkItem, error)
	DiscoverWorkflow(ctx *config.Context) (map[string]config.WorkflowMap, error)
	TransitionWorkItems(ctx *config.Context, items []*config.CacheItem, ref config.TransitionRef) error
}

func NewProviderClient(p *config.Provider) (ProviderClient, error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	switch p.Type {
	case "jira":
		return &JiraProviderClient{
			cfg:    p,
			client: &client,
		}, nil
	case "azure":
		return &AzureProviderClient{
			cfg:    p,
			client: &client,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", p.Type)
	}
}
