package azure

import (
	"net/url"
	"strconv"
	"strings"
)

type AzureEndpoints struct {
	apiVersion   string
	host         string
	organization string
	project      string
}

func host2url(host string) *url.URL {
	u, err := url.Parse(host)
	if err != nil {
		panic(err)
	}
	return u
}

func (o *AzureEndpoints) testEndpoint() string {
	u := host2url(o.host)
	u = u.JoinPath(o.organization, "_apis", "projects")

	q := url.Values{}
	q.Set("api-version", o.apiVersion)

	u.RawQuery = q.Encode()
	return u.String()
}

func (o *AzureEndpoints) connectionDataEndpoint() string {
	u := host2url(o.host)
	u = u.JoinPath(o.organization, "_apis", "ConnectionData")
	return u.String()
}

func (o *AzureEndpoints) wiqlEndpoint() string {
	u := host2url(o.host)
	u = u.JoinPath(o.organization, o.project, "_apis", "wit", "wiql")

	q := url.Values{}
	q.Set("api-version", o.apiVersion)

	u.RawQuery = q.Encode()
	return u.String()
}

func (o *AzureEndpoints) workItemsEndpoint(ids []string) string {
	u := host2url(o.host)
	u = u.JoinPath(o.organization, o.project, "_apis", "wit", "workitems")

	q := url.Values{}
	q.Set("ids", strings.Join(ids, ","))
	q.Set("api-version", o.apiVersion)

	u.RawQuery = q.Encode()
	return u.String()
}

func (o *AzureEndpoints) workItemEndpoint(id int) string {
	u := host2url(o.host)
	u = u.JoinPath(o.organization, o.project, "_apis", "wit", "workitems", strconv.Itoa(id))

	q := url.Values{}
	q.Set("api-version", o.apiVersion)

	u.RawQuery = q.Encode()
	return u.String()
}

func (o *AzureEndpoints) workItemDetailEndpoint(id int) string {
	u := host2url(o.host)
	u = u.JoinPath(o.organization, o.project, "_apis", "wit", "workitems", strconv.Itoa(id))

	q := url.Values{}
	q.Set("api-version", o.apiVersion)
	q.Set("fields", strings.Join([]string{
		"System.Title",
		"System.Description",
		"System.CreatedDate",
		"System.CreatedBy",
		"System.IterationPath",
		"System.AssignedTo",
	}, ","))

	u.RawQuery = q.Encode()
	return u.String()
}

func (o *AzureEndpoints) workItemCommentsEndpoint(id int) string {
	u := host2url(o.host)
	u = u.JoinPath(o.organization, o.project, "_apis", "wit", "workitems", strconv.Itoa(id), "comments")

	q := url.Values{}
	q.Set("api-version", o.apiVersion+"-preview")

	u.RawQuery = q.Encode()
	return u.String()
}

func (o *AzureEndpoints) workItemCommentsListEndpoint(id int, top int) string {
	u := host2url(o.host)
	u = u.JoinPath(o.organization, o.project, "_apis", "wit", "workitems", strconv.Itoa(id), "comments")

	q := url.Values{}
	q.Set("api-version", o.apiVersion+"-preview")
	q.Set("$top", strconv.Itoa(top))
	q.Set("order", "desc")

	u.RawQuery = q.Encode()
	return u.String()
}

func (o *AzureEndpoints) workItemWebEndpoint(id int) string {
	u := host2url(o.host)
	u = u.JoinPath(o.organization, o.project, "_workitems", "edit", strconv.Itoa(id))
	return u.String()
}

// assignEndpoint reuses workItemEndpoint: Azure has no dedicated assign URL,
// assignment is a field update (PATCH) on the work item.
func (o *AzureEndpoints) assignEndpoint(id int) string {
	return o.workItemEndpoint(id)
}

func (o *AzureEndpoints) workItemTypesEndpoint() string {
	u := host2url(o.host)
	u = u.JoinPath(o.organization, o.project, "_apis", "wit", "workitemtypes")

	q := url.Values{}
	q.Set("api-version", o.apiVersion)

	u.RawQuery = q.Encode()
	return u.String()
}

func (o *AzureEndpoints) workItemTypeStatesEndpoint(workItemType string) string {
	u := host2url(o.host)
	u = u.JoinPath(o.organization, o.project, "_apis", "wit", "workitemtypes", workItemType, "states")

	q := url.Values{}
	q.Set("api-version", o.apiVersion)

	u.RawQuery = q.Encode()
	return u.String()
}
