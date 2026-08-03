package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
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

// jiraStatusCategory.Key is locale-independent ("new", "indeterminate", "done");
// Name is localized and only used for display.
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

	issues, err := c.searchRaw(jql, fields)
	if err != nil {
		return nil, err
	}

	items := make([]WorkItem, 0, len(issues))
	for _, issue := range issues {
		items = append(items, jiraIssueToWorkItem(issue))
	}
	return items, nil
}

// searchRaw fetches all pages of a search query and returns the raw issues.
func (c *JiraProviderClient) searchRaw(
	jql, fields string,
) ([]jiraIssue, error) {
	var all []jiraIssue
	var nextPageToken string

	for {
		issues, token, err := c.queryApi(jql, fields, nextPageToken)
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
	jql string, fields string,
	nextPageToken string,
) ([]jiraIssue, string, error) {

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

	var token string
	if !search.IsLast {
		token = search.NextPageToken
	}

	return search.Issues, token, nil
}

func (c *JiraProviderClient) getTransitions(
	issueKey string,
) ([]jiraTransition, error) {
	apiURL := c.cfg.Host +
		"/rest/api/3/issue/" + url.PathEscape(issueKey) + "/transitions"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.cfg.Username, c.cfg.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var transitions jiraTransitionsResponse
	if err := json.Unmarshal(respBody, &transitions); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return transitions.Transitions, nil
}

// TransitionWorkItems moves the given work items into the target state
// described by ref (e.g. WorkflowMap.Start to mark them "in progress").
// Jira transitions are dynamic per issue, so ref must carry the transition
// key (the numeric transition ID) that was resolved for the work items'
// issue type. ctx is only present to satisfy the ProviderClient interface
// and is not used by Jira.
func (c *JiraProviderClient) TransitionWorkItems(
	_ *config.Context, items []*config.CacheItem, ref config.TransitionRef,
) error {
	if ref.TransitionKey == "" {
		return fmt.Errorf(
			"cannot transition work items: no transition key set (target state %q)",
			ref.DisplayName,
		)
	}

	var failures []string
	for _, item := range items {
		if item.Status == ref.DisplayName {
			// TODO Inform user, or ignore?
			continue
		}

		if err := c.transition(item.Key, ref.TransitionKey); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.Key, err))
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf(
			"failed to transition %d of %d work items:\n  %s",
			len(failures), len(items), strings.Join(failures, "\n  "),
		)
	}

	return nil
}

func (c *JiraProviderClient) transition(issueKey, transitionID string) error {
	apiURL := c.cfg.Host +
		"/rest/api/3/issue/" + url.PathEscape(issueKey) + "/transitions"

	payload, err := json.Marshal(jiraTransitionRequest{
		Transition: jiraTransitionID{ID: transitionID},
	})
	if err != nil {
		return fmt.Errorf("bad request body: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.cfg.Username, c.cfg.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}

// DiscoverWorkflow samples the most prevalent issue types in the project and
// resolves their start/done transitions by matching the status category of
// each transition's target status.
func (c *JiraProviderClient) DiscoverWorkflow(
	ctx *config.Context,
) (map[string]config.WorkflowMap, error) {

	project := ctx.Remote.Project
	if project == "" {
		return nil, fmt.Errorf("jira project is required in context")
	}

	jql := fmt.Sprintf("project = %s ORDER BY updated DESC", project)

	// We use up to one page of work items to read transitions from.
	issues, _, err := c.queryApi(jql, "summary,status,issuetype", "")
	if err != nil {
		return nil, err
	}

	types, issuesByType := groupByPrevalentTypes(issues, 5)

	workflow := make(map[string]config.WorkflowMap)
	for _, t := range types {
		wm, err := c.discoverWorkflowForType(issuesByType[t])
		if err != nil {
			return nil, err
		}
		workflow[t] = wm
	}

	// The most prevalent type serves as the default for unmapped types
	// before runtime exploration will overwrite transitions for the unknwon
	// type later on
	if len(types) > 0 {
		workflow["default"] = workflow[types[0]]
	}

	return workflow, nil
}

func (c *JiraProviderClient) discoverWorkflowForType(
	issues []jiraIssue,
) (config.WorkflowMap, error) {

	var wm config.WorkflowMap
	var transitionErr error

	for _, issue := range issues {
		transitions, err := c.getTransitions(issue.Key)
		if err != nil {
			transitionErr = err
			continue
		}

		for _, tr := range transitions {
			switch tr.To.StatusCategory.Key {
			case "indeterminate":
				if wm.Start.TransitionKey == "" {
					wm.Start = config.TransitionRef{
						TransitionKey: tr.ID,
						DisplayName:   tr.To.Name,
					}
				}
			case "done":
				if wm.Done.TransitionKey == "" {
					wm.Done = config.TransitionRef{
						TransitionKey: tr.ID,
						DisplayName:   tr.To.Name,
					}
				}
			}
		}

		if wm.Start.TransitionKey != "" && wm.Done.TransitionKey != "" {
			break
		}
	}

	if wm.Start.TransitionKey == "" && wm.Done.TransitionKey == "" && transitionErr != nil {
		return wm, transitionErr
	}
	return wm, nil
}

// groupByPrevalentTypes groups issues by issue type and returns the type names
// sorted by frequency (descending), capped at maxTypes.
func groupByPrevalentTypes(
	issues []jiraIssue, maxTypes int,
) ([]string, map[string][]jiraIssue) {

	counts := make(map[string]int)
	byType := make(map[string][]jiraIssue)
	for _, issue := range issues {
		t := issue.Fields.IssType.Name
		if t == "" {
			continue
		}
		counts[t]++
		byType[t] = append(byType[t], issue)
	}

	types := make([]string, 0, len(counts))
	for t := range counts {
		types = append(types, t)
	}
	sort.SliceStable(types, func(i, j int) bool {
		if counts[types[i]] != counts[types[j]] {
			return counts[types[i]] > counts[types[j]]
		}
		return types[i] < types[j]
	})

	// TODO: Keep this or remove it as it feels arbitrary?
	if len(types) > maxTypes {
		types = types[:maxTypes]
	}

	return types, byType
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
