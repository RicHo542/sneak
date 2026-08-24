package client

import (
	"context"
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
	GetUserIdent() (*objects.UserInfo, error)
	ListWorkItems(context.Context, *config.LocalContext, objects.ListOptions) ([]objects.WorkItem, error)
	DiscoverWorkflow(context.Context, *config.LocalContext) (map[string]config.WorkflowMap, error)
	DiscoverWorkflowForItem(context.Context, *config.CacheItem) (config.WorkflowMap, error)
	TransitionWorkItems(context.Context, *config.LocalContext, []*config.CacheItem, config.TransitionRef) error
	StartWorkItems(context.Context, *config.LocalContext, []*config.CacheItem, config.TransitionRef) error
	AddCommentToWorkItems(context.Context, *config.LocalContext, []*config.CacheItem, string) error
	AssignWorkItems(context.Context, *config.LocalContext, []*config.CacheItem) error
	DescribeWorkItem(context.Context, string) (*objects.WorkItemDetail, error)
}

func NewProviderClient(lctx *config.LocalContext, p *config.Provider) (ProviderClient, error) {
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
			p, &client, lctx,
		), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", p.Type)
	}
}
