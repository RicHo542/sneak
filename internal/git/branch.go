package git

import (
	"strings"

	"github.com/richo542/sneak/internal/config"
)

func BuildBranchName(tasks []*config.CacheItem) string {
	var normalizedNames []string

	for _, t := range tasks {
		normalizedNames = append(normalizedNames, strings.ToLower(t.Key))
	}

	return "feat/" + strings.Join(normalizedNames, "-")
}
