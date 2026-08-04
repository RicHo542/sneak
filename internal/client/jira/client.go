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

type jiraSearchResponse struct {
	IsLast        bool        `json:"isLast"`
	NextPageToken string      `json:"nextPageToken"`
	Issues        []jiraIssue `json:"issues"`
}

type jiraIssue struct {
	ID     string     `json:"id"`
	Key    string     `json:"key"`
	Fields jiraFields `json:"fields"`
}

type jiraFields struct {
	Summary  string         `json:"summary"`
	Status   jiraStatus     `json:"status"`
	IssType  jiraNamedValue `json:"issuetype"`
	Assignee *jiraUserValue `json:"assignee"`
}

type jiraNamedValue struct {
	Name string `json:"name"`
}

type jiraStatus struct {
	Name           string             `json:"name"`
	StatusCategory jiraStatusCategory `json:"statusCategory"`
}

type jiraStatusCategory struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type jiraUserValue struct {
	Name string `json:"displayName"`
}

type jiraTransitionsResponse struct {
	Transitions []jiraTransition `json:"transitions"`
}

type jiraTransition struct {
	ID   string     `json:"id"`
	Name string     `json:"name"`
	To   jiraStatus `json:"to"`
}

type jiraTransitionRequest struct {
	Transition jiraTransitionID `json:"transition"`
}

type jiraTransitionID struct {
	ID string `json:"id"`
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
