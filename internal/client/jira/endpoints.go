package jira

import (
	"fmt"
	"net/url"
)

type JiraEndpoints struct {
	apiVersion string
	host       string
}

func (o *JiraEndpoints) myselfEndpoint() string {
	return fmt.Sprintf(
		"%s/rest/api/%s/myself",
		o.host, o.apiVersion,
	)
}

func (o *JiraEndpoints) testEndpoint() string {
	return o.myselfEndpoint()
}

func (o *JiraEndpoints) queryEndpoint() string {
	return fmt.Sprintf(
		"%s/rest/api/%s/search/jql",
		o.host, o.apiVersion,
	)
}

func (o *JiraEndpoints) transitionEndpoint(issueKey string) string {
	return fmt.Sprintf(
		"%s/rest/api/%s/issue/%s/transitions",
		o.host, o.apiVersion,
		url.PathEscape(issueKey),
	)
}

func (o *JiraEndpoints) commentEndpoint(issueKey string) string {
	return fmt.Sprintf(
		"%s/rest/api/%s/issue/%s/comment",
		o.host, o.apiVersion,
		url.PathEscape(issueKey),
	)
}
