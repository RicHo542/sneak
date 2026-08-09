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

// AssignWorkItems assigns the given work items to the authenticated user,
// whose assignable handle is stored in the provider config (UserHandle).
func (c *AzureProviderClient) AssignWorkItems(
	ctx context.Context, lctx *config.LocalContext,
	items []*config.CacheItem,
) error {

	project := lctx.Remote.Project
	if project == "" {
		return fmt.Errorf("azure project is required in context")
	}

	payload, err := json.Marshal([]azureWorkItemPatch{{
		Op:    "add",
		Path:  "/fields/System.AssignedTo",
		Value: c.Cfg.UserHandle,
	}})
	if err != nil {
		return fmt.Errorf("bad request body: %w", err)
	}

	var fails []string
	for _, item := range items {
		id, err := azureWorkItemID(item.Key)
		if err != nil {
			fails = append(fails, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}

		apiURL := c.Endpoints.assignEndpoint(project, id)

		req, err := http.NewRequestWithContext(ctx, "PATCH", apiURL, bytes.NewReader(payload))
		if err != nil {
			fails = append(fails, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}
		c.SetAuthHeader(req)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json-patch+json")

		resp, err := c.Client.Do(req)
		if err != nil {
			fails = append(fails, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fails = append(fails, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fails = append(fails, fmt.Sprintf(
				"%s: HTTP %d: %s", item.Key, resp.StatusCode,
				strings.TrimSpace(string(respBody)),
			))
		}

		// Keep cache up to date
		item.Assignee = c.Cfg.UserDisplayName
	}

	if len(fails) > 0 {
		return fmt.Errorf(
			"failed to assign %d of %d work items:\n  %s",
			len(fails), len(items), strings.Join(fails, "\n  "),
		)
	}

	return nil
}
