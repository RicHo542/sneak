package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/richo542/sneak/internal/client/objects"
	"github.com/richo542/sneak/internal/config"
)

func CheckAndRefreshCache(app *App, refresh bool) (bool, error) {
	state := app.State
	bindingsChanged := !state.Cache.MatchesBindings(app.Context.Bindings)
	needsFetch := refresh || bindingsChanged || !state.Cache.IsFresh()

	if needsFetch {
		if err := RefreshCache(app); err != nil {
			return needsFetch, err
		}
	}

	return needsFetch, nil
}

func RefreshCache(app *App) error {
	fmt.Println("Fetching work items...")
	state := app.State

	items, err := app.Client.ListWorkItems(app.Context, objects.ListOptions{
		Bindings: app.Context.Bindings,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch work items: %w", err)
	}

	cacheItems := make([]config.CacheItem, len(items))
	for i, item := range items {
		cacheItems[i] = config.CacheItem{
			ID:       item.ID,
			Key:      item.Key,
			Summary:  item.Summary,
			Status:   item.Status,
			Type:     item.Type,
			Assignee: item.Assignee,
		}
	}

	state.Cache = config.Cache{
		Items:     cacheItems,
		FetchedAt: time.Now(),
		Bindings:  app.Context.Bindings,
	}

	if err := config.SaveState(app.Dir, state); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	return nil
}

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

func NormalizeTaskNames(tasks []string) []string {
	var normalizedNames []string

	for _, t := range tasks {
		normalizedNames = append(normalizedNames, strings.ToLower(t))
	}

	return normalizedNames
}

func GroupTasksByTransition(
	app *App, tasks []*config.CacheItem, action string,
) (map[string]transitionGroup, error) {
	groups := make(map[string]transitionGroup)
	for _, t := range tasks {
		workflow, err := ResolveTransitionForTask(app, t, action, 0)
		if err != nil {
			return nil, err
		}

		group := groups[workflow.Start.TransitionKey]
		if group.ref.TransitionKey == "" {
			group.ref = workflow.Start
		}
		group.items = append(group.items, t)
		groups[workflow.Start.TransitionKey] = group
	}
	return groups, nil
}

func ResolveTransitionForTask(
	app *App, task *config.CacheItem, action string, attempt int,
) (*config.WorkflowMap, error) {

	workflow, err := app.Context.GetWorkflowByType(task.Type)

	var focusedKey string
	if workflow != nil {
		focusedKey = workflow.Start.TransitionKey
		if action == "close" {
			focusedKey = workflow.Done.TransitionKey
		}
	}

	// Attempt to auto-discover the workflow once for the new task type.
	// In case the workflow transition cannot be found, error out to the user.
	if err != nil || focusedKey == "" {
		discovered, discoverErr := app.Client.DiscoverWorkflowForItem(task)
		if discoverErr == nil {
			app.Context.SetWorkflowForTaskType(app.Dir, task.Type, discovered)
			if attempt < 1 {
				return ResolveTransitionForTask(
					app, task, action, attempt+1,
				)
			}
		}
		if discoverErr == nil {
			discoverErr = fmt.Errorf(
				"discovered workflow has no transition for action '%s'", action,
			)
		}
		return nil, fmt.Errorf(
			"unable to find transition key for action "+
				"'%s' and type '%s': %w", action, task.Type, discoverErr,
		)
	}

	return workflow, nil
}

func TransitionCacheItems(
	app *App, cacheItems []*config.CacheItem, action string,
) error {
	// Move tasks to in progress
	// Identify the transition target for each task by its type
	// Build groups with transitionKey -> work items
	groups, err := GroupTasksByTransition(app, cacheItems, action)
	if err != nil {
		return err
	}

	for _, group := range groups {
		if err := app.Client.TransitionWorkItems(
			app.Context, group.items,
			group.ref,
		); err != nil {
			return fmt.Errorf(
				"failed to move work items to new state "+
					"for action '%s': %w", action, err,
			)
		}

		// Keeps the cache updated and makes sure
		// active_tasks will also get the right state
		for _, item := range group.items {
			item.Status = group.ref.DisplayName
		}
	}

	return nil
}
