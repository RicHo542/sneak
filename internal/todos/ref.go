package todos

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/richo542/sneak/internal/config"
)

const (
	TodoStatusOpen         = "open"
	TodoStatusActive       = "active"
	TodoStatusClosed       = "done"
	TodoKeyPrefix          = "sn:"
	TodoGlobalScopeKeyword = "global"
	TodoFilename           = "todos.json"
)

var ValidStatus = map[string]bool{
	"open":   true,
	"active": true,
	"done":   true,
}

func IsValidTodoStatus(status string) bool {
	_, ok := ValidStatus[status]
	return ok
}

func HasTodoRef(val string) bool {
	return strings.HasPrefix(val, TodoKeyPrefix)
}

func TodoFilePath() (string, error) {
	configDir, err := config.SneakConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, TodoFilename), nil
}

func SortForDisplay(items []*Todo) {
	slices.SortStableFunc(items, func(a, b *Todo) int {
		if a.Pin != b.Pin {
			if a.Pin {
				return -1
			}
			return 1
		}
		return a.CreatedAt.Compare(b.CreatedAt)
	})
}
