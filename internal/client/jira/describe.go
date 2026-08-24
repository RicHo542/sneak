package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/richo542/sneak/internal/client/objects"
)

const describeCommentsLimit = 5

type jiraNamedUser struct {
	DisplayName string `json:"displayName"`
}

type jiraIssueDetailFields struct {
	Summary     string         `json:"summary"`
	Description jiraADFNode    `json:"description"`
	Created     string         `json:"created"`
	Creator     *jiraNamedUser `json:"creator"`
	Assignee    *jiraNamedUser `json:"assignee"`
}

type jiraIssueDetailResponse struct {
	Fields jiraIssueDetailFields `json:"fields"`
}

type jiraCommentEntry struct {
	Author  *jiraNamedUser `json:"author"`
	Body    jiraADFNode    `json:"body"`
	Created string         `json:"created"`
}

type jiraCommentsResponse struct {
	Total    int                `json:"total"`
	Comments []jiraCommentEntry `json:"comments"`
}

type jiraSprintInfo struct {
	Name string `json:"name"`
}

type jiraSprintFields struct {
	Sprint *jiraSprintInfo `json:"sprint"`
}

type jiraSprintResponse struct {
	Fields jiraSprintFields `json:"fields"`
}

// DescribeWorkItem fetches a full detail view (fields + recent comments)
// for a single issue. Always live; never cached.
func (c *JiraProviderClient) DescribeWorkItem(
	ctx context.Context, key string,
) (*objects.WorkItemDetail, error) {
	fields, err := c.getIssueDetailFields(ctx, key)
	if err != nil {
		return nil, err
	}

	comments, total, err := c.getRecentComments(ctx, key, describeCommentsLimit)
	if err != nil {
		return nil, err
	}

	detail := &objects.WorkItemDetail{
		ID:            key,
		Key:           key,
		URL:           c.Endpoints.browseEndpoint(key),
		Name:          fields.Summary,
		Description:   adfToText(fields.Description),
		CreatedAt:     fields.Created,
		IterationPath: c.getSprint(ctx, key),
		Comments:      comments,
		TotalComments: total,
	}
	if fields.Creator != nil {
		detail.CreatedBy = fields.Creator.DisplayName
	}
	if fields.Assignee != nil {
		detail.Owner = fields.Assignee.DisplayName
	}

	return detail, nil
}

func (c *JiraProviderClient) getIssueDetailFields(
	ctx context.Context, key string,
) (*jiraIssueDetailFields, error) {
	apiURL := c.Endpoints.issueDetailEndpoint(key)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.Cfg.Username, c.Cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
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

	var parsed jiraIssueDetailResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse issue: %w", err)
	}

	return &parsed.Fields, nil
}

func (c *JiraProviderClient) getRecentComments(
	ctx context.Context, key string, limit int,
) ([]objects.Comment, int, error) {
	apiURL := c.Endpoints.commentsListEndpoint(key, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(c.Cfg.Username, c.Cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed jiraCommentsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, 0, fmt.Errorf("failed to parse comments: %w", err)
	}

	// orderBy=-created returns most-recent-first; reverse to oldest-first
	// for display.
	comments := make([]objects.Comment, len(parsed.Comments))
	for i, cm := range parsed.Comments {
		out := objects.Comment{
			Body:      adfToText(cm.Body),
			CreatedAt: cm.Created,
		}
		if cm.Author != nil {
			out.Author = cm.Author.DisplayName
		}
		comments[len(parsed.Comments)-1-i] = out
	}

	return comments, parsed.Total, nil
}

// getSprint resolves the issue's current sprint/iteration name via the
// Agile REST API. This is best-effort: instances without Jira Software (no
// boards/sprints), or issues with no sprint set, will fall back to
// "(unknown)" rather than failing the whole describe call.
func (c *JiraProviderClient) getSprint(ctx context.Context, key string) string {
	apiURL := c.Endpoints.sprintEndpoint(key)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "(unknown)"
	}
	req.SetBasicAuth(c.Cfg.Username, c.Cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "(unknown)"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "(unknown)"
	}

	var parsed jiraSprintResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "(unknown)"
	}

	if parsed.Fields.Sprint == nil || parsed.Fields.Sprint.Name == "" {
		return "(unknown)"
	}
	return parsed.Fields.Sprint.Name
}
