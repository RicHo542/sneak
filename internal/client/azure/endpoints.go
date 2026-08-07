package azure

import (
	"fmt"
	"net/url"
	"strings"
)

type AzureEndpoints struct {
	apiVersion string
	host       string
}

func (o *AzureEndpoints) testEndpoint() string {
	return fmt.Sprintf("%s/_apis", o.host)
}

func (o *AzureEndpoints) connectionDataEndpoint() string {
	return fmt.Sprintf(
		"%s/_apis/ConnectionData?api-version=%s",
		o.host, o.apiVersion,
	)
}

func (o *AzureEndpoints) wiqlEndpoint(project string) string {
	return fmt.Sprintf(
		"%s/%s/_apis/wit/wiql?api-version=%s",
		o.host, url.PathEscape(project), o.apiVersion,
	)
}

func (o *AzureEndpoints) workItemsEndpoint(project string, ids []string) string {
	return fmt.Sprintf(
		"%s/%s/_apis/wit/workitems?ids=%s&api-version=%s",
		o.host, url.PathEscape(project), strings.Join(ids, ","), o.apiVersion,
	)
}

func (o *AzureEndpoints) workItemEndpoint(project string, id int) string {
	return fmt.Sprintf(
		"%s/%s/_apis/wit/workitems/%d?api-version=%s",
		o.host, url.PathEscape(project), id, o.apiVersion,
	)
}

func (o *AzureEndpoints) workItemCommentsEndpoint(project string, id int) string {
	return fmt.Sprintf(
		"%s/%s/_apis/wit/workitems/%d/comments?api-version=%s",
		o.host, url.PathEscape(project), id, o.apiVersion,
	)
}

func (o *AzureEndpoints) workItemTypesEndpoint(project string) string {
	return fmt.Sprintf(
		"%s/%s/_apis/wit/workitemtypes?api-version=%s",
		o.host, url.PathEscape(project), o.apiVersion,
	)
}

func (o *AzureEndpoints) workItemTypeStatesEndpoint(project, workItemType string) string {
	return fmt.Sprintf(
		"%s/%s/_apis/wit/workitemtypes/%s/states?api-version=%s",
		o.host, url.PathEscape(project), url.PathEscape(workItemType), o.apiVersion,
	)
}
