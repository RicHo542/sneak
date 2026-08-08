package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/richo542/sneak/internal/client/objects"
	"github.com/richo542/sneak/internal/config"
)

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

func (c *JiraProviderClient) ListWorkItems(
	ctx context.Context, lctx *config.LocalContext, opts objects.ListOptions,
) ([]objects.WorkItem, error) {

	if len(opts.Bindings) == 0 {
		return nil, fmt.Errorf("no bindings specified. Please specify bindings in 'sneak init'.")
	}

	return c.listRecursive(ctx, lctx, opts.Bindings, opts.Types)
}

func (c *JiraProviderClient) listRecursive(
	ctx context.Context, lctx *config.LocalContext,
	bindings []string, types []string,
) ([]objects.WorkItem, error) {

	seen := make(map[string]bool)
	var all []objects.WorkItem
	current := bindings

	// Recursivley go through children relations until the last
	// children do not have children anymore.
	// For each child, all childs will be queried recursivley through
	// all returned API pages to  ensure also long lists of children
	// are parsed properly.
	for len(current) > 0 {
		jql := AssignedChildrenJql(lctx.Remote.Project, current, types)
		children, err := c.queryApiRecursive(ctx, jql, "summary,status,issuetype,assignee")
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
	ctx context.Context, jql string, fields string,
) ([]objects.WorkItem, error) {

	issues, err := c.searchRaw(ctx, jql, fields)
	if err != nil {
		return nil, err
	}

	items := make([]objects.WorkItem, 0, len(issues))
	for _, issue := range issues {
		items = append(items, jiraIssueToWorkItem(issue))
	}
	return items, nil
}

// searchRaw fetches all pages of a search query and returns the raw issues.
func (c *JiraProviderClient) searchRaw(
	ctx context.Context, jql string, fields string,
) ([]jiraIssue, error) {
	var all []jiraIssue
	var nextPageToken string

	for {
		issues, token, err := c.queryApi(ctx, jql, fields, nextPageToken)
		if err != nil {
			return nil, err
		}

		all = append(all, issues...)

		// Last page - Exit
		if token == "" {
			break
		}
		nextPageToken = token
	}

	return all, nil
}

func (c *JiraProviderClient) queryApi(
	ctx context.Context, jql string,
	fields string, nextPageToken string,
) ([]jiraIssue, string, error) {

	params := url.Values{}
	params.Set("jql", jql)
	params.Set("fields", fields)
	params.Set("expand", "")

	if nextPageToken != "" {
		params.Set("nextPageToken", nextPageToken)
	}

	apiURL := c.Endpoints.queryEndpoint() + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.Cfg.Username, c.Cfg.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US")

	resp, err := c.Client.Do(req)
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

	var token string
	if !search.IsLast {
		token = search.NextPageToken
	}

	return search.Issues, token, nil
}
