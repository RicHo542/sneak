package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/richo542/sneak/internal/config"
)

// StartWorkItems assigns the given work items to the authenticated user and
// moves them into the target start state described by ref in a single PATCH.
// Assigning and transitioning in one request avoids Azure's optimistic
// concurrency (TF26071) that concurrent writes to the same work item trigger.
func (c *AzureProviderClient) StartWorkItems(
	ctx context.Context, lctx *config.LocalContext,
	items []*config.CacheItem, ref config.TransitionRef,
) error {
	if ref.TransitionKey == "" {
		return fmt.Errorf("cannot start work items: no target state set")
	}

	project := lctx.Remote.Project
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

		ops := []azureWorkItemPatch{{
			Op:    "add",
			Path:  "/fields/System.AssignedTo",
			Value: c.Cfg.UserHandle,
		}}
		if item.Status != ref.DisplayName {
			ops = append(ops, azureWorkItemPatch{
				Op:    "add",
				Path:  "/fields/System.State",
				Value: ref.TransitionKey,
			})
		}

		if err := c.patchWorkItem(ctx, id, ops); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}

		item.Assignee = c.Cfg.UserDisplayName
		item.Status = ref.DisplayName
	}

	if len(failures) > 0 {
		return fmt.Errorf(
			"failed to start %d of %d work items:\n  %s",
			len(failures), len(items), strings.Join(failures, "\n  "),
		)
	}
	return nil
}

func (c *AzureProviderClient) patchWorkItem(
	ctx context.Context, id int, ops []azureWorkItemPatch,
) error {
	apiURL := c.Endpoints.workItemEndpoint(id)

	payload, err := json.Marshal(ops)
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
