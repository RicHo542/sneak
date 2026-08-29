package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/git"
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

func ResolveCloseTaskFocus(app *App, taskKeys []string, all bool) ([]*config.CacheItem, error) {
	return resolveTaskFocus(app, taskKeys, all, app.State.GetActiveCacheItems)
}

func ResolveShipTaskFocus(app *App, taskKeys []string, all bool) ([]*config.CacheItem, error) {
	return resolveTaskFocus(app, taskKeys, all, app.State.GetActiveCacheItems)
}

func ResolveStartTaskFocus(app *App, taskKeys []string, all bool) ([]*config.CacheItem, error) {
	return resolveTaskFocus(app, taskKeys, all, app.State.GetNonActiveCacheItems)
}

func resolveTaskFocus(
	app *App, taskKeys []string, all bool,
	candidateFunc func() ([]*config.CacheItem, error),
) ([]*config.CacheItem, error) {

	if len(taskKeys) > 0 {
		cacheItems, err := app.State.Cache.GetByKeyBatch(taskKeys)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to find tasks in cache. "+
					"Consider to run 'sneak list --refresh' to force a refresh.: %w",
				err,
			)
		}
		return cacheItems, nil
	}

	cacheItems, err := candidateFunc()
	if err != nil {
		return nil, err
	}

	if all {
		return cacheItems, nil
	}

	// Prompt user with interactive selection if no key was provided.
	selection, tuiSelectErr := ui.InteractiveSelectItem(cacheItems)
	if tuiSelectErr != nil {
		return nil, fmt.Errorf("failed task selection.")
	}
	return selection, nil
}

func DiscoverGitRepos(dir string, multirepo bool) ([]string, error) {
	gc := git.NewGitClient()
	if gc.IsRepo(dir) {
		return []string{dir}, nil
	}

	if !multirepo {
		return []string{}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s: %w", dir, err)
	}

	var repos []string

	for _, candidate := range entries {
		if !candidate.IsDir() {
			continue
		}

		candidatePath := filepath.Join(dir, candidate.Name())
		if gc.IsRepo(candidatePath) {
			repos = append(repos, candidatePath)
		}
	}

	return repos, nil
}
