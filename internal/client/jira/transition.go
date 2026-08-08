package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/richo542/sneak/internal/config"
)

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

func (c *JiraProviderClient) getTransitions(
	ctx context.Context, issueKey string,
) ([]jiraTransition, error) {
	apiURL := c.Endpoints.transitionEndpoint(issueKey)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.Cfg.Username, c.Cfg.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US")

	resp, err := c.Client.Do(req)
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
// and is not used by Jira, but required by Azure.
func (c *JiraProviderClient) TransitionWorkItems(
	ctx context.Context, _ *config.LocalContext,
	items []*config.CacheItem, ref config.TransitionRef,
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

		if err := c.transition(ctx, item.Key, ref.TransitionKey); err != nil {
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

func (c *JiraProviderClient) transition(
	ctx context.Context, issueKey string,
	transitionID string,
) error {
	apiURL := c.Cfg.Host +
		"/rest/api/3/issue/" + url.PathEscape(issueKey) + "/transitions"

	payload, err := json.Marshal(jiraTransitionRequest{
		Transition: jiraTransitionID{ID: transitionID},
	})
	if err != nil {
		return fmt.Errorf("bad request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.Cfg.Username, c.Cfg.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
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
	ctx context.Context, lctx *config.LocalContext,
) (map[string]config.WorkflowMap, error) {

	project := lctx.Remote.Project
	if project == "" {
		return nil, fmt.Errorf("jira project is required in context")
	}

	jql := fmt.Sprintf("project = %s ORDER BY updated DESC", project)

	// We use up to one page of work items to read transitions from.
	issues, _, err := c.queryApi(ctx, jql, "summary,status,issuetype", "")
	if err != nil {
		return nil, err
	}

	types, issuesByType := groupByPrevalentTypes(issues, 5)

	workflow := make(map[string]config.WorkflowMap)
	for _, t := range types {
		wm, err := c.discoverWorkflowForIssues(ctx, issuesByType[t])
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

func (c *JiraProviderClient) discoverWorkflowForIssues(
	ctx context.Context, issues []jiraIssue,
) (config.WorkflowMap, error) {

	var wm config.WorkflowMap
	var transitionErr error

	for _, issue := range issues {
		transitions, err := c.getTransitions(ctx, issue.Key)
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

func (c *JiraProviderClient) DiscoverWorkflowForItem(
	ctx context.Context, task *config.CacheItem,
) (config.WorkflowMap, error) {

	var wm config.WorkflowMap
	transitions, err := c.getTransitions(ctx, task.Key)
	if err != nil {
		return wm, err
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

		if wm.Start.TransitionKey != "" && wm.Done.TransitionKey != "" {
			break
		}
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
