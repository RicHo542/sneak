package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/richo542/sneak/internal/config"
)

type JiraProviderClient struct {
	cfg    *config.Provider
	client *http.Client
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
	Status   jiraNamedValue `json:"status"`
	IssType  jiraNamedValue `json:"issuetype"`
	Assignee *jiraUserValue `json:"assignee"`
}

type jiraNamedValue struct {
	Name string `json:"name"`
}

type jiraUserValue struct {
	Name string `json:"displayName"`
}

func (c *JiraProviderClient) TestConnection() error {
	host := strings.TrimRight(c.cfg.Host, "/")
	apiURL := host + "/rest/api/3/myself"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.cfg.Username, c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	return TestRequest(c.client, req)
}

func (c *JiraProviderClient) ListWorkItems(
	ctx *config.Context, opts ListOptions,
) ([]WorkItem, error) {

	if len(opts.Bindings) == 0 {
		return nil, fmt.Errorf("no bindings specified. Please specify bindings in 'sneak init'.")
	}

	return c.listRecursive(ctx, opts.Bindings, opts.Types)
}

func (c *JiraProviderClient) listRecursive(
	ctx *config.Context, bindings []string,
	types []string,
) ([]WorkItem, error) {

	seen := make(map[string]bool)
	var all []WorkItem
	current := bindings

	// Recursivley go through children relations until the last
	// children do not have children anymore.
	// For each child, all childs will be queried recursivley through
	// all returned API pages to  ensure also long lists of children
	// are parsed properly.
	for len(current) > 0 {
		jql := AssignedChildrenJql(ctx.Remote.Project, current, types)
		children, err := c.queryApiRecursive(jql, "summary,status,issuetype,assignee")
		if err != nil {
			return nil, err
		}

		var next []string
		for _, child := range children {
			if !seen[child.Key] {
				seen[child.Key] = true
				all = append(all, child)
				next = append(next, child.Key)
			}
		}

		current = next
	}

	return all, nil
}

func (c *JiraProviderClient) queryApiRecursive(
	jql string, fields string,
) ([]WorkItem, error) {
	var all []WorkItem
	var nextPageToken string

	for {
		items, token, err := c.queryApi(jql, fields, nextPageToken)
		if err != nil {
			return nil, err
		}

		all = append(all, items...)

		// Last page - Exit
		if token == "" {
			break
		}
		nextPageToken = token
	}

	return all, nil
}

func (c *JiraProviderClient) queryApi(
	jql string, fields string,
	nextPageToken string,
) ([]WorkItem, string, error) {

	params := url.Values{}
	params.Set("jql", jql)
	params.Set("fields", fields)
	params.Set("expand", "")

	if nextPageToken != "" {
		params.Set("nextPageToken", nextPageToken)
	}

	apiURL := c.cfg.Host + "/rest/api/3/search/jql?" + params.Encode()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.cfg.Username, c.cfg.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var search jiraSearchResponse
	if err := json.Unmarshal(respBody, &search); err != nil {
		return nil, "", fmt.Errorf("failed to parse response: %w", err)
	}

	var items []WorkItem
	for _, issue := range search.Issues {
		items = append(items, jiraIssueToWorkItem(issue))
	}

	var token string
	if !search.IsLast {
		token = search.NextPageToken
	}

	return items, token, nil
}

func jiraIssueToWorkItem(issue jiraIssue) WorkItem {
	wi := WorkItem{
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
