package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/richo542/sneak/internal/config"
)

type azureCommentRequest struct {
	Text string `json:"text"`
}

// AddCommentToWorkItems posts the given comment to each of the passed work
// items.
func (c *AzureProviderClient) AddCommentToWorkItems(
	ctx *config.Context, items []*config.CacheItem,
	comment string,
) error {
	comment = strings.TrimSpace(comment)

	project := ctx.Remote.Project
	if project == "" {
		return fmt.Errorf("azure project is required in context")
	}

	payload, err := json.Marshal(azureCommentRequest{Text: comment})
	if err != nil {
		return fmt.Errorf("bad request body: %w", err)
	}

	var failures []string
	for _, item := range items {
		id, err := azureWorkItemID(item.Key)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}

		apiURL := c.Endpoints.workItemCommentsEndpoint(project, id)

		req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payload))
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}

		c.SetAuthHeader(req)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.Client.Do(req)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			failures = append(failures, fmt.Sprintf(
				"%s: HTTP %d: %s", item.Key, resp.StatusCode,
				strings.TrimSpace(string(respBody)),
			))
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf(
			"failed to comment on %d of %d work items:\n  %s",
			len(failures), len(items), strings.Join(failures, "\n  "),
		)
	}

	return nil
}
