package jira

import (
	"net/url"
	"strconv"
	"strings"
)

type JiraEndpoints struct {
	apiVersion string
	host       string
}

func host2url(host string) *url.URL {
	u, err := url.Parse(host)
	if err != nil {
		panic(err)
	}
	return u
}

func (o *JiraEndpoints) myselfEndpoint() string {
	u := host2url(o.host)
	u = u.JoinPath("rest", "api", o.apiVersion, "myself")
	return u.String()
}

func (o *JiraEndpoints) testEndpoint() string {
	return o.myselfEndpoint()
}

func (o *JiraEndpoints) queryEndpoint() string {
	u := host2url(o.host)
	u = u.JoinPath("rest", "api", o.apiVersion, "search", "jql")
	return u.String()
}

func (o *JiraEndpoints) transitionEndpoint(issueKey string) string {
	u := host2url(o.host)
	u = u.JoinPath("rest", "api", o.apiVersion, "issue", issueKey, "transitions")
	return u.String()
}

func (o *JiraEndpoints) commentEndpoint(issueKey string) string {
	u := host2url(o.host)
	u = u.JoinPath("rest", "api", o.apiVersion, "issue", issueKey, "comment")
	return u.String()
}

func (o *JiraEndpoints) commentsListEndpoint(issueKey string, maxResults int) string {
	u := host2url(o.host)
	u = u.JoinPath("rest", "api", o.apiVersion, "issue", issueKey, "comment")

	q := url.Values{}
	q.Set("orderBy", "-created")
	q.Set("maxResults", strconv.Itoa(maxResults))

	u.RawQuery = q.Encode()
	return u.String()
}

func (o *JiraEndpoints) issueDetailEndpoint(issueKey string) string {
	u := host2url(o.host)
	u = u.JoinPath("rest", "api", o.apiVersion, "issue", issueKey)

	q := url.Values{}
	q.Set("fields", strings.Join([]string{
		"summary", "description", "created", "creator", "assignee",
	}, ","))

	u.RawQuery = q.Encode()
	return u.String()
}

func (o *JiraEndpoints) sprintEndpoint(issueKey string) string {
	u := host2url(o.host)
	u = u.JoinPath("rest", "agile", "1.0", "issue", issueKey)

	q := url.Values{}
	q.Set("fields", "sprint")

	u.RawQuery = q.Encode()
	return u.String()
}

func (o *JiraEndpoints) browseEndpoint(issueKey string) string {
	u := host2url(o.host)
	u = u.JoinPath("browse", issueKey)
	return u.String()
}

func (o *JiraEndpoints) assignEndpoint(issueKey string) string {
	u := host2url(o.host)
	u = u.JoinPath("rest", "api", o.apiVersion, "issue", issueKey, "assignee")
	return u.String()
}
