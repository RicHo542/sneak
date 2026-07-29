package git

import "strings"

// Continue
func BuildBranchName(tasks []string) string {
	// var normalizedNames []string

	for _, t := range tasks {
		strings.ToLower(t)
	}

	return ""
}
