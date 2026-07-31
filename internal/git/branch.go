package git

import (
	"strings"
)

func BuildBranchName(tasks []string) string {
	var normalizedNames []string

	for _, t := range tasks {
		normalizedNames = append(normalizedNames, strings.ToLower(t))
	}

	return strings.Join(normalizedNames, "-")
}
