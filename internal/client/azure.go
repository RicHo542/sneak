package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/richo542/sneak/internal/config"
)

type AzureProviderClient struct {
	cfg    *config.Provider
	client *http.Client
}

type azureWIQLRequest struct {
	Query string `json:"query"`
}

type azureWIQLResponse struct {
	WorkItems []azureWIQLWorkItem `json:"workItems"`
}

type azureWIQLWorkItem struct {
	ID int `json:"id"`
}

type azureBatchResponse []azureBatchWorkItem

type azureBatchWorkItem struct {
	ID     int              `json:"id"`
	Fields azureBatchFields `json:"fields"`
}

type azureBatchFields struct {
	Title    string         `json:"System.Title"`
	Type     string         `json:"System.WorkItemType"`
	State    string         `json:"System.State"`
	Assignee *azureAssignee `json:"System.AssignedTo"`
}

type azureAssignee struct {
	DisplayName string `json:"displayName"`
}

type azureWorkItemTypesResponse struct {
	Value []azureWorkItemType `json:"value"`
}

type azureWorkItemType struct {
	Name string `json:"name"`
}

type azureStatesResponse struct {
	Value []azureWorkItemState `json:"value"`
}

// azureWorkItemState.StateCategory is language-independent
// ("Proposed", "InProgress", "Resolved", "Completed").
type azureWorkItemState struct {
	Name          string `json:"name"`
	StateCategory string `json:"stateCategory"`
}

type azureWorkItemPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func (c *AzureProviderClient) TestConnection() error {
	host := strings.TrimRight(c.cfg.Host, "/")
	apiURL := host + "/_apis/profile/profiles/me?api-version=7.0"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth("", c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	return TestRequest(c.client, req)
}

func (c *AzureProviderClient) ListWorkItems(ctx *config.Context, opts ListOptions) ([]WorkItem, error) {
	host := strings.TrimRight(c.cfg.Host, "/")

	project := ctx.Remote.Project
	if project == "" {
		return nil, fmt.Errorf("azure project is required in context")
	}

	wiqlQuery := c.buildWIQL(ctx, opts)
	ids, err := c.queryWIQL(host, project, wiqlQuery)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []WorkItem{}, nil
	}

	return c.fetchBatch(host, project, ids)
}

func (c *AzureProviderClient) queryWIQL(host, project, query string) ([]int, error) {
	apiURL := fmt.Sprintf("%s/%s/_apis/wit/wiql?api-version=7.0", host, project)

	body, err := json.Marshal(azureWIQLRequest{Query: query})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal WIQL: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth("", c.cfg.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var wiqlResp azureWIQLResponse
	if err := json.Unmarshal(respBody, &wiqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse WIQL response: %w", err)
	}

	ids := make([]int, 0, len(wiqlResp.WorkItems))
	for _, wi := range wiqlResp.WorkItems {
		ids = append(ids, wi.ID)
	}
	return ids, nil
}

func (c *AzureProviderClient) fetchBatch(host, project string, ids []int) ([]WorkItem, error) {
	var all []WorkItem
	for i := 0; i < len(ids); i += 200 {
		end := i + 200
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[i:end]

		items, err := c.fetchChunk(host, project, chunk)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
	}
	return all, nil
}

func (c *AzureProviderClient) fetchChunk(host, project string, ids []int) ([]WorkItem, error) {
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = strconv.Itoa(id)
	}
	apiURL := fmt.Sprintf("%s/%s/_apis/wit/workitems?ids=%s&$expand=relations&api-version=7.0",
		host, project, strings.Join(idStrs, ","))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth("", c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var batch azureBatchResponse
	if err := json.Unmarshal(body, &batch); err != nil {
		return nil, fmt.Errorf("failed to parse batch response: %w", err)
	}

	var items []WorkItem
	for _, wi := range batch {
		item := WorkItem{
			ID:      strconv.Itoa(wi.ID),
			Key:     fmt.Sprintf("#%d", wi.ID),
			Summary: wi.Fields.Title,
			Status:  wi.Fields.State,
			Type:    wi.Fields.Type,
		}
		if wi.Fields.Assignee != nil {
			item.Assignee = wi.Fields.Assignee.DisplayName
		}
		items = append(items, item)
	}
	return items, nil
}

func (c *AzureProviderClient) buildWIQL(ctx *config.Context, opts ListOptions) string {
	project := ctx.Remote.Project

	if len(opts.Bindings) > 0 {
		return c.buildRecursiveWIQL(ctx, opts)
	}

	var conditions []string
	conditions = append(conditions, fmt.Sprintf("[System.TeamProject] = '%s'", project))

	if ctx.Remote.Team != "" {
		conditions = append(conditions, fmt.Sprintf("[System.AreaPath] = '%s\\%s'", project, ctx.Remote.Team))
	}

	if len(opts.Types) != 0 {
		conditions = append(conditions, fmt.Sprintf("[System.WorkItemType] IN ('%s')", strings.Join(opts.Types, ", ")))
	}

	query := "SELECT [System.Id], [System.WorkItemType], [System.Title], [System.State], [System.AssignedTo] FROM WorkItems"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY [System.ChangedDate] DESC"

	return query
}

func (c *AzureProviderClient) buildRecursiveWIQL(ctx *config.Context, opts ListOptions) string {
	project := ctx.Remote.Project

	var where []string
	where = append(where, fmt.Sprintf("[System.TeamProject] = '%s'", project))
	where = append(where, "[System.Links.LinkType] = 'System.LinkTypes.Hierarchy-Forward'")
	where = append(where, fmt.Sprintf("[Source].[System.Id] IN (%s)", strings.Join(opts.Bindings, ", ")))

	if len(opts.Types) != 0 {
		where = append(
			where,
			fmt.Sprintf("[Target].[System.WorkItemType] IN ('%s')", strings.Join(opts.Types, ", ")),
		)
	}

	query := fmt.Sprintf(
		"SELECT [System.Id], [System.WorkItemType], [System.Title], [System.State], [System.AssignedTo] FROM WorkItemLinks WHERE %s MODE (Recursive) ORDER BY [System.ChangedDate] DESC",
		strings.Join(where, " AND "),
	)

	return query
}

// DiscoverWorkflow resolves start/done states per work item type from the
// static state definitions (no issue sampling required).
func (c *AzureProviderClient) DiscoverWorkflow(
	ctx *config.Context,
) (map[string]config.WorkflowMap, error) {

	host := strings.TrimRight(c.cfg.Host, "/")
	project := ctx.Remote.Project
	if project == "" {
		return nil, fmt.Errorf("azure project is required in context")
	}

	types, err := c.getWorkItemTypes(host, project)
	if err != nil {
		return nil, err
	}

	workflow := make(map[string]config.WorkflowMap)
	for _, t := range types {
		wm, err := c.getWorkflowForType(host, project, t)
		if err != nil {
			return nil, err
		}
		workflow[t] = wm
	}

	if len(types) > 0 {
		workflow["default"] = workflow[types[0]]
	}

	return workflow, nil
}

// TransitionWorkItems moves the given work items into the target state
// described by ref. Azure transitions are direct state changes, so only
// ref.TransitionKey (the target state name) is used; work item keys may be
// given as "#123" or plain numeric IDs.
func (c *AzureProviderClient) TransitionWorkItems(
	ctx *config.Context, items []*config.CacheItem, ref config.TransitionRef,
) error {
	if ref.TransitionKey == "" {
		return fmt.Errorf("cannot transition work items: no target state set")
	}

	host := strings.TrimRight(c.cfg.Host, "/")
	project := ctx.Remote.Project
	if project == "" {
		return fmt.Errorf("azure project is required in context")
	}

	var failures []string
	for _, item := range items {
		id, err := azureWorkItemID(item.Key)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}
		if err := c.transition(host, project, id, ref.TransitionKey); err != nil {
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

// azureWorkItemID converts a work item key (e.g. "#123" or "123") to its
// numeric ID.
func azureWorkItemID(key string) (int, error) {
	idStr := strings.TrimPrefix(strings.TrimSpace(key), "#")
	if idStr == "" {
		return 0, fmt.Errorf("invalid work item key %q", key)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid work item key %q", key)
	}
	return id, nil
}

func (c *AzureProviderClient) transition(
	host, project string, id int, state string,
) error {
	apiURL := fmt.Sprintf(
		"%s/%s/_apis/wit/workitems/%d?api-version=7.0",
		host, project, id,
	)

	payload, err := json.Marshal([]azureWorkItemPatch{{
		Op:    "add",
		Path:  "/fields/System.State",
		Value: state,
	}})
	if err != nil {
		return fmt.Errorf("bad request body: %w", err)
	}

	req, err := http.NewRequest("PATCH", apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth("", c.cfg.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json-patch+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}

func (c *AzureProviderClient) getWorkItemTypes(host, project string) ([]string, error) {
	apiURL := fmt.Sprintf("%s/%s/_apis/wit/workitemtypes?api-version=7.0", host, project)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth("", c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed azureWorkItemTypesResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse work item types: %w", err)
	}

	types := make([]string, 0, len(parsed.Value))
	for _, t := range parsed.Value {
		types = append(types, t.Name)
	}
	return types, nil
}

func (c *AzureProviderClient) getWorkItemTypeStates(
	host, project, workItemType string,
) ([]azureWorkItemState, error) {

	apiURL := fmt.Sprintf(
		"%s/%s/_apis/wit/workitemtypes/%s/states?api-version=7.0",
		host, project, url.PathEscape(workItemType),
	)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth("", c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed azureStatesResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse states: %w", err)
	}

	return parsed.Value, nil
}

func (c *AzureProviderClient) getWorkflowForType(
	host, project, workItemType string,
) (config.WorkflowMap, error) {

	states, err := c.getWorkItemTypeStates(host, project, workItemType)
	if err != nil {
		return config.WorkflowMap{}, err
	}

	var wm config.WorkflowMap
	for _, s := range states {
		if s.StateCategory == "InProgress" && wm.Start.TransitionKey == "" {
			wm.Start = config.TransitionRef{TransitionKey: s.Name, DisplayName: s.Name}
		}
		if s.StateCategory == "Completed" && wm.Done.TransitionKey == "" {
			wm.Done = config.TransitionRef{TransitionKey: s.Name, DisplayName: s.Name}
		}
	}

	// Some types (e.g. Bug) resolve through "Resolved" before "Completed".
	if wm.Done.TransitionKey == "" {
		for _, s := range states {
			if s.StateCategory == "Resolved" {
				wm.Done = config.TransitionRef{TransitionKey: s.Name, DisplayName: s.Name}
				break
			}
		}
	}

	return wm, nil
}
