package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/richo542/sneak/internal/config"
)

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
	Name     string `json:"name"`
	Category string `json:"category"`
}

type azureWorkItemPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

// TransitionWorkItems moves the given work items into the target state
// described by ref. Azure transitions are direct state changes, so only
// ref.TransitionKey (the target state name) is used; work item keys may be
// given as "#123" or plain numeric IDs.
func (c *AzureProviderClient) TransitionWorkItems(
	ctx context.Context, lctx *config.LocalContext,
	items []*config.CacheItem, ref config.TransitionRef,
) error {
	if ref.TransitionKey == "" {
		return fmt.Errorf("cannot transition work items: no target state set")
	}

	project := lctx.Remote.Project
	if project == "" {
		return fmt.Errorf("azure project is required in context")
	}

	var failures []string
	for _, item := range items {
		if item.Status == ref.DisplayName {
			continue
		}

		id, err := azureWorkItemID(item.Key)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}
		if err := c.transition(ctx, id, ref.TransitionKey); err != nil {
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
	idStr := strings.TrimSpace(key)
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
	ctx context.Context, id int, state string,
) error {
	apiURL := c.Endpoints.workItemEndpoint(id)

	payload, err := json.Marshal([]azureWorkItemPatch{{
		Op:    "add",
		Path:  "/fields/System.State",
		Value: state,
	}})
	if err != nil {
		return fmt.Errorf("bad request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	c.SetAuthHeader(req)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json-patch+json")

	resp, err := c.Client.Do(req)
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

// DiscoverWorkflow resolves start/done states per work item type from the
// static state definitions (no issue sampling required).
func (c *AzureProviderClient) DiscoverWorkflow(
	ctx context.Context, lctx *config.LocalContext,
) (map[string]config.WorkflowMap, error) {

	project := lctx.Remote.Project
	if project == "" {
		return nil, fmt.Errorf("azure project is required in context")
	}

	types, err := c.getWorkItemTypes(ctx, project)
	if err != nil {
		return nil, err
	}

	workflow := make(map[string]config.WorkflowMap)
	for _, t := range types {
		wm, err := c.getWorkflowForType(ctx, t)
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

func (c *AzureProviderClient) getWorkItemTypes(
	ctx context.Context, project string,
) ([]string, error) {
	apiURL := c.Endpoints.workItemTypesEndpoint()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}
	c.SetAuthHeader(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
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
	ctx context.Context, workItemType string,
) ([]azureWorkItemState, error) {

	apiURL := c.Endpoints.workItemTypeStatesEndpoint(workItemType)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}
	c.SetAuthHeader(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
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
	ctx context.Context, workItemType string,
) (config.WorkflowMap, error) {

	states, err := c.getWorkItemTypeStates(ctx, workItemType)
	if err != nil {
		return config.WorkflowMap{}, err
	}

	var wm config.WorkflowMap
	for _, s := range states {
		if s.Category == "InProgress" && wm.Start.TransitionKey == "" {
			wm.Start = config.TransitionRef{TransitionKey: s.Name, DisplayName: s.Name}
		}
		if s.Category == "Completed" && wm.Done.TransitionKey == "" {
			wm.Done = config.TransitionRef{TransitionKey: s.Name, DisplayName: s.Name}
		}
	}

	// Some types (e.g. Bug) resolve through "Resolved" before "Completed".
	if wm.Done.TransitionKey == "" {
		for _, s := range states {
			if s.Category == "Resolved" {
				wm.Done = config.TransitionRef{TransitionKey: s.Name, DisplayName: s.Name}
				break
			}
		}
	}

	return wm, nil
}

// DiscoverWorkflowForItem resolves start/done states for a single work item.
func (c *AzureProviderClient) DiscoverWorkflowForItem(
	ctx context.Context, task *config.CacheItem,
) (config.WorkflowMap, error) {
	project := c.Endpoints.project
	if project == "" {
		return config.WorkflowMap{}, fmt.Errorf(
			"azure project is required, cannot discover workflow for item %q", task.Key,
		)
	}
	if task.Type == "" {
		return config.WorkflowMap{}, fmt.Errorf(
			"cannot discover workflow for item %q: work item type is unknown", task.Key,
		)
	}
	return c.getWorkflowForType(ctx, task.Type)
}
