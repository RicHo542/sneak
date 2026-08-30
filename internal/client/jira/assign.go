package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/richo542/sneak/internal/config"
)

type jiraAssignRequest struct {
	AccountId string `json:"accountId"`
}

// UnassignWorkItems clears the assignee of the given work items. Jira clears
// the assignee when the request body is an empty object; sending an empty
// accountId instead would be interpreted as assigning to a non-existent user.
func (c *JiraProviderClient) UnassignWorkItems(ctx context.Context, _ *config.LocalContext, items []*config.CacheItem) error {
	payload := []byte("{}")

	var fails []string
	for _, item := range items {
		url := c.Endpoints.assignEndpoint(item.Key)

		req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(payload))
		if err != nil {
			fails = append(fails, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}
		req.SetBasicAuth(c.Cfg.Username, c.Cfg.Token)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

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
		item.Assignee = ""
	}

	if len(fails) > 0 {
		return fmt.Errorf(
			"failed to unassign %d of %d work items:\n  %s",
			len(fails), len(items), strings.Join(fails, "\n  "),
		)
	}

	return nil
}

// StartWorkItems assigns the given work items to the authenticated user and
// moves them into the target start state described by ref. Jira has no single
// combined request for both, so the two operations are performed sequentially.
func (c *JiraProviderClient) StartWorkItems(
	ctx context.Context, lctx *config.LocalContext,
	items []*config.CacheItem, ref config.TransitionRef,
) error {
	return errors.Join(
		c.AssignWorkItems(ctx, lctx, items),
		c.TransitionWorkItems(ctx, lctx, items, ref),
	)
}

func (c *JiraProviderClient) AssignWorkItems(ctx context.Context, _ *config.LocalContext, items []*config.CacheItem) error {

	payload, err := json.Marshal(jiraAssignRequest{
		AccountId: c.Cfg.UserHandle,
	})
	if err != nil {
		return fmt.Errorf("bad request body: %w", err)
	}

	var fails []string
	for _, item := range items {
		url := c.Endpoints.assignEndpoint(item.Key)

		req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(payload))
		if err != nil {
			fails = append(fails, fmt.Sprintf("%s: %v", item.Key, err))
			continue
		}
		req.SetBasicAuth(c.Cfg.Username, c.Cfg.Token)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

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
