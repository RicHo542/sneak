package azure

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

type azureWorkItemDetailResponse struct {
	ID     int                       `json:"id"`
	Fields azureWorkItemDetailFields `json:"fields"`
}

type azureWorkItemDetailFields struct {
	Title         string         `json:"System.Title"`
	Description   string         `json:"System.Description"`
	CreatedDate   string         `json:"System.CreatedDate"`
	CreatedBy     *azureAssignee `json:"System.CreatedBy"`
	IterationPath string         `json:"System.IterationPath"`
	AssignedTo    *azureAssignee `json:"System.AssignedTo"`
}

type azureCommentsListResponse struct {
	TotalCount int            `json:"totalCount"`
	Comments   []azureComment `json:"comments"`
}

type azureComment struct {
	Text        string         `json:"text"`
	CreatedBy   *azureAssignee `json:"createdBy"`
	CreatedDate string         `json:"createdDate"`
}

// DescribeWorkItem fetches a full detail view (fields + recent comments)
// for a single work item. Always live; never cached.
func (c *AzureProviderClient) DescribeWorkItem(
	ctx context.Context, key string,
) (*objects.WorkItemDetail, error) {
	id, err := azureWorkItemID(key)
	if err != nil {
		return nil, err
	}

	fields, err := c.getWorkItemDetailFields(ctx, id)
	if err != nil {
		return nil, err
	}

	comments, total, err := c.getRecentComments(ctx, id, describeCommentsLimit)
	if err != nil {
		return nil, err
	}

	detail := &objects.WorkItemDetail{
		ID:            fmt.Sprintf("%d", id),
		Key:           fmt.Sprintf("%d", id),
		URL:           c.Endpoints.workItemWebEndpoint(id),
		Name:          fields.Title,
		Description:   stripHTML(fields.Description),
		CreatedAt:     fields.CreatedDate,
		IterationPath: fields.IterationPath,
		Comments:      comments,
		TotalComments: total,
	}
	if fields.CreatedBy != nil {
		detail.CreatedBy = fields.CreatedBy.DisplayName
	}
	if fields.AssignedTo != nil {
		detail.Owner = fields.AssignedTo.DisplayName
	}

	return detail, nil
}

func (c *AzureProviderClient) getWorkItemDetailFields(
	ctx context.Context, id int,
) (*azureWorkItemDetailFields, error) {
	apiURL := c.Endpoints.workItemDetailEndpoint(id)

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed azureWorkItemDetailResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse work item: %w", err)
	}

	return &parsed.Fields, nil
}

func (c *AzureProviderClient) getRecentComments(
	ctx context.Context, id int, limit int,
) ([]objects.Comment, int, error) {
	apiURL := c.Endpoints.workItemCommentsListEndpoint(id, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("bad request: %w", err)
	}
	c.SetAuthHeader(req)
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

	var parsed azureCommentsListResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, 0, fmt.Errorf("failed to parse comments: %w", err)
	}

	// Response is most-recent-first (order=desc); reverse to oldest-first
	// for display.
	comments := make([]objects.Comment, len(parsed.Comments))
	for i, cm := range parsed.Comments {
		out := objects.Comment{
			Body:      stripHTML(cm.Text),
			CreatedAt: cm.CreatedDate,
		}
		if cm.CreatedBy != nil {
			out.Author = cm.CreatedBy.DisplayName
		}
		comments[len(parsed.Comments)-1-i] = out
	}

	return comments, parsed.TotalCount, nil
}
