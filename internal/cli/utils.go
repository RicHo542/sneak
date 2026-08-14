package cli

import (
	"fmt"

	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/ui"
)

func HasInvalidTasks(app *App, tasks []string) ([]string, bool) {
	// validate that task names are actually referring to
	// existing tasks
	var invalidTasks []string
	for _, task := range tasks {
		valid := false
		for _, item := range app.State.Cache.Items {
			if task == item.Key {
				valid = true
			}
		}

		if !valid {
			invalidTasks = append(invalidTasks, task)
		}
	}

	return invalidTasks, len(invalidTasks) > 0
}

func ResolveTaskFocus(app *App, taskKeys []string, all bool) ([]*config.CacheItem, error) {
	if all {
		cacheItems, err := app.State.GetActiveCacheItems()
		if err != nil {
			return nil, err
		}
		return cacheItems, nil
	}

	selectedKeys := taskKeys
	tuiSelectErr := fmt.Errorf("failed task selection.")
	// Prompt user with interactive selection if no key was provided.
	if len(taskKeys) < 1 {
		selectedKeys, tuiSelectErr = ui.InteractiveSelectItem(app.State.Cache.Items)
		if tuiSelectErr != nil {
			return nil, tuiSelectErr
		}
	}

	cacheItems, err := app.State.Cache.GetByKeyBatch(selectedKeys)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to find tasks in cache. "+
				"Consider to run 'sneak list --refresh' to force a refresh.: %w",
			err,
		)
	}

	return cacheItems, nil
}
