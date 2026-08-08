package jira

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

// jiraADFNode is a node in the Atlassian Document Format (ADF) tree used to
// describe a comment body.
type jiraADFNode struct {
	Type    string        `json:"type"`
	Version int           `json:"version,omitempty"`
	Text    string        `json:"text,omitempty"`
	Content []jiraADFNode `json:"content,omitempty"`
}

type jiraCommentRequest struct {
	Body jiraADFNode `json:"body"`
}

// AddCommentToWorkItems posts the given comment to each of
// the passed work item.
func (c *JiraProviderClient) AddCommentToWorkItems(
	ctx context.Context, _ *config.LocalContext,
	items []*config.CacheItem, comment string,
) error {
	comment = strings.TrimSpace(comment)

	payload, err := json.Marshal(jiraCommentRequest{
		Body: jiraADFNode{
			Type:    "doc",
			Version: 1,
			Content: []jiraADFNode{{
				Type: "paragraph",
				Content: []jiraADFNode{{
					Type: "text",
					Text: comment,
				}},
			}},
		},
	})
	if err != nil {
		return fmt.Errorf("bad request body: %w", err)
	}

	var failures []string
	for _, item := range items {
		apiURL := c.Endpoints.commentEndpoint(item.Key)

		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payload))
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}

		req.SetBasicAuth(c.Cfg.Username, c.Cfg.Token)
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
