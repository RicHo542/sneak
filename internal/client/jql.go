package client

import (
	"fmt"
	"strings"
)

func AssignedChildrenJql(
	project string, parents []string,
	types []string,
) string {

	jql := fmt.Sprintf("project=%s", project)

	if len(parents) > 0 {
		parentsClause := fmt.Sprintf(
			"parent in (%s)", strings.Join(parents, ", "),
		)
		jql += fmt.Sprintf(" AND %s", parentsClause)
	}

	if len(types) > 0 {
		typesClause := fmt.Sprintf(
			"issuetype in (\"%s\")", strings.Join(types, ", "),
		)
		jql += fmt.Sprintf(" AND %s", typesClause)

	}

	jql += " AND (assignee = currentUser() OR assignee IS EMPTY)"
	jql += " order by created DESC"

	return jql
}
