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

// CoarseStatus maps any provider or local status string onto one of the unified
// categories ("open", "active", "done"). Returns "" for unknown statuses.
func CoarseStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case TodoStatusOpen, TodoStatusActive, TodoStatusClosed:
		return strings.ToLower(status)
	case "new", "proposed", "todo", "to do", "to-do", "backlog",
		"untriaged", "planned", "unstarted", "not started", "created",
		"ready":
		return TodoStatusOpen
	case "in progress", "in-progress", "doing", "committed", "approved",
		"in review", "in-review", "review", "testing", "in testing",
		"waiting", "blocked", "started", "in development":
		return TodoStatusActive
	case "closed", "resolved", "completed", "removed", "cancelled",
		"canceled", "accepted", "rejected", "duplicate", "wontfix",
		"won't fix", "expired":
		return TodoStatusClosed
	default:
		return ""
	}
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
