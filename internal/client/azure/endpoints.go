package azure

import (
	"fmt"
	"net/url"
	"strings"
)

type AzureEndpoints struct {
	apiVersion   string
	host         string
	organization string
	project      string
}

func (o *AzureEndpoints) testEndpoint() string {
	return fmt.Sprintf(
		"%s/%s/_apis/projects?api-version=%s",
		o.host, o.organization, o.apiVersion,
	)
}

func (o *AzureEndpoints) connectionDataEndpoint() string {
	return fmt.Sprintf(
		"%s/%s/_apis/ConnectionData",
		o.host, o.organization,
	)
}

func (o *AzureEndpoints) wiqlEndpoint() string {
	return fmt.Sprintf(
		"%s/%s/%s/_apis/wit/wiql?api-version=%s",
		o.host, o.organization, o.project, o.apiVersion,
	)
}

func (o *AzureEndpoints) workItemsEndpoint(ids []string) string {
	return fmt.Sprintf(
		"%s/%s/%s/_apis/wit/workitems?ids=%s&api-version=%s",
		o.host, o.organization, o.project, strings.Join(ids, ","), o.apiVersion,
	)
}

func (o *AzureEndpoints) workItemEndpoint(id int) string {
	return fmt.Sprintf(
		"%s/%s/%s/_apis/wit/workitems/%d?api-version=%s",
		o.host, o.organization, o.project, id, o.apiVersion,
	)
}

func (o *AzureEndpoints) workItemCommentsEndpoint(project string, id int) string {
	return fmt.Sprintf(
		"%s/%s/%s/_apis/wit/workitems/%d/comments?api-version=%s",
		o.host, o.organization, url.PathEscape(project), id, o.apiVersion,
	)
}

// assignEndpoint reuses workItemEndpoint: Azure has no dedicated assign URL,
// assignment is a field update (PATCH) on the work item.
func (o *AzureEndpoints) assignEndpoint(project string, id int) string {
	return o.workItemEndpoint(id)
}

func (o *AzureEndpoints) workItemTypesEndpoint() string {
	return fmt.Sprintf(
		"%s/%s/%s/_apis/wit/workitemtypes?api-version=%s",
		o.host, o.organization, o.project, o.apiVersion,
	)
}

func (o *AzureEndpoints) workItemTypeStatesEndpoint(workItemType string) string {
	return fmt.Sprintf(
		"%s/%s/%s/_apis/wit/workitemtypes/%s/states?api-version=%s",
		o.host, o.organization, o.project, url.PathEscape(workItemType), o.apiVersion,
	)
}
