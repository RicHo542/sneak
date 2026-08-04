package jira

import (
	"net/http"
	"strings"

	"github.com/richo542/sneak/internal/client/objects"
	"github.com/richo542/sneak/internal/config"
)

type JiraProviderClient struct {
	Cfg       *config.Provider
	Client    *http.Client
	Endpoints *JiraEndpoints
}

func NewJiraProviderClient(
	cfg *config.Provider, client *http.Client,
) *JiraProviderClient {
	host := strings.TrimRight(cfg.Host, "/")

	return &JiraProviderClient{
		Cfg:       cfg,
		Client:    client,
		Endpoints: &JiraEndpoints{host: host, apiVersion: "3"},
	}
}

func jiraIssueToWorkItem(issue jiraIssue) objects.WorkItem {
	wi := objects.WorkItem{
		ID:      issue.ID,
		Key:     issue.Key,
		Summary: issue.Fields.Summary,
		Status:  issue.Fields.Status.Name,
		Type:    issue.Fields.IssType.Name,
	}
	if issue.Fields.Assignee != nil {
		wi.Assignee = issue.Fields.Assignee.Name
	}
	return wi
}
