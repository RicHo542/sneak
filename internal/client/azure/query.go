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

	"github.com/richo542/sneak/internal/client/objects"
	"github.com/richo542/sneak/internal/config"
)

type azureWIQLRequest struct {
	Query string `json:"query"`
}

type azureWIQLResponse struct {
	WorkItems         []azureWIQLWorkItem `json:"workItems"`
	WorkItemRelations []azureWIQLRelation `json:"workItemRelations"`
}

type azureWIQLWorkItem struct {
	ID int `json:"id"`
}

type azureWIQLRelation struct {
	Rel    string            `json:"rel"`
	Source azureWIQLWorkItem `json:"source"`
	Target azureWIQLWorkItem `json:"target"`
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

func (c *AzureProviderClient) ListWorkItems(
	ctx context.Context, lctx *config.LocalContext, opts objects.ListOptions,
) ([]objects.WorkItem, error) {
	project := lctx.Remote.Project
	if project == "" {
		return nil, fmt.Errorf("azure project is required in context")
	}

	wiqlQuery := c.buildWIQL(lctx, opts)
	ids, err := c.queryWIQL(ctx, project, wiqlQuery)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []objects.WorkItem{}, nil
	}

	return c.fetchBatch(ctx, project, ids)
}

func (c *AzureProviderClient) queryWIQL(
	ctx context.Context, project string, query string,
) ([]int, error) {
	apiURL := c.Endpoints.wiqlEndpoint(project)

	body, err := json.Marshal(azureWIQLRequest{Query: query})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal WIQL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}
	c.SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")
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

	var wiqlResp azureWIQLResponse
	if err := json.Unmarshal(respBody, &wiqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse WIQL response: %w", err)
	}

	// Flat queries return "workItems"; tree/link queries (MODE Recursive)
	// return "workItemRelations", whose targets are the matched descendants.
	var ids []int
	if len(wiqlResp.WorkItemRelations) > 0 {
		for _, rel := range wiqlResp.WorkItemRelations {
			ids = append(ids, rel.Target.ID)
		}
	} else {
		for _, wi := range wiqlResp.WorkItems {
			ids = append(ids, wi.ID)
		}
	}
	return ids, nil
}

func (c *AzureProviderClient) fetchBatch(
	ctx context.Context, project string, ids []int,
) ([]objects.WorkItem, error) {
	var all []objects.WorkItem
	for i := 0; i < len(ids); i += 200 {
		end := i + 200
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[i:end]

		items, err := c.fetchChunk(ctx, project, chunk)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
	}
	return all, nil
}

func (c *AzureProviderClient) fetchChunk(
	ctx context.Context, project string, ids []int,
) ([]objects.WorkItem, error) {
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = strconv.Itoa(id)
	}
	apiURL := c.Endpoints.workItemsEndpoint(project, idStrs)

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

	var batch azureBatchResponse
	if err := json.Unmarshal(body, &batch); err != nil {
		return nil, fmt.Errorf("failed to parse batch response: %w", err)
	}

	items := make([]objects.WorkItem, 0, len(batch))
	for _, wi := range batch {
		items = append(items, azureWorkItemToObject(wi))
	}
	return items, nil
}

func (c *AzureProviderClient) buildWIQL(
	ctx *config.LocalContext, opts objects.ListOptions,
) string {
	if len(opts.Bindings) > 0 {
		return c.buildRecursiveWIQL(ctx, opts)
	}

	project := ctx.Remote.Project

	var conditions []string
	conditions = append(conditions, fmt.Sprintf("[System.TeamProject] = '%s'", project))
	conditions = append(conditions, "[System.AssignedTo] = @me")

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

func (c *AzureProviderClient) buildRecursiveWIQL(
	ctx *config.LocalContext, opts objects.ListOptions,
) string {
	project := ctx.Remote.Project

	var where []string
	where = append(where, fmt.Sprintf("[Source].[System.TeamProject] = '%s'", project))
	where = append(where, "[System.Links.LinkType] = 'System.LinkTypes.Hierarchy-Forward'")
	where = append(where, fmt.Sprintf("[Source].[System.Id] IN (%s)", strings.Join(opts.Bindings, ", ")))

	if len(opts.Types) != 0 {
		where = append(
			where,
			fmt.Sprintf("[Target].[System.WorkItemType] IN ('%s')", strings.Join(opts.Types, ", ")),
		)
	}

	// Only work items assigned to the current user are returned, mirroring
	// the Jira provider's "assignee = currentUser() OR assignee IS EMPTY".
	where = append(where, "[Target].[System.AssignedTo] = @me")

	query := fmt.Sprintf(
		"SELECT [System.Id] FROM WorkItemLinks WHERE %s MODE (Recursive)",
		strings.Join(where, " AND "),
	)

	return query
}
