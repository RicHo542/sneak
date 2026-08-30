package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/ui"
)

// transitionGroup groups work items by the transition key that moves them
// into their target state, so each group needs only one transition call.
type transitionGroup struct {
	ref   config.TransitionRef
	items []*config.CacheItem
}

func CommentCacheItems(app *app.App, cacheItems []*config.CacheItem, comment string) {
	comment = strings.TrimSpace(comment)
	if comment != "" {
		commentErr := app.Client.AddCommentToWorkItems(
			app.Ctx, app.LocalContext, cacheItems, comment,
		)
		if commentErr != nil {
			ui.Printfln("failed to add comment: %v", commentErr)
		}
	}
}

func CloseCacheItems(
	app *app.App, cacheItems []*config.CacheItem,
) error {
	return transitionCacheItems(
		app, cacheItems, "close", app.Client.TransitionWorkItems,
	)
}

func StartCacheItems(
	app *app.App, cacheItems []*config.CacheItem,
) error {
	return transitionCacheItems(
		app, cacheItems, "start", app.Client.StartWorkItems,
	)
}

// ReopenCacheItems moves work items back to the open (to-do) state.
func ReopenCacheItems(
	app *app.App, cacheItems []*config.CacheItem,
) error {
	return transitionCacheItems(
		app, cacheItems, "reopen", app.Client.TransitionWorkItems,
	)
}

func transitionCacheItems(
	app *app.App, cacheItems []*config.CacheItem, action string,
	exec func(context.Context, *config.LocalContext, []*config.CacheItem, config.TransitionRef) error,
) error {
	// Move tasks to a new transition state (in progress / closed)
	// Identify the transition target for each task by its type
	// Build groups with transitionKey -> work items
	groups, err := GroupTasksByTransition(app, cacheItems, action)
	if err != nil {
		return err
	}

	for _, group := range groups {
		if err := exec(
			app.Ctx, app.LocalContext, group.items,
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

func GroupTasksByTransition(
	app *app.App, tasks []*config.CacheItem, action string,
) (map[string]transitionGroup, error) {
	groups := make(map[string]transitionGroup)
	for _, t := range tasks {
		workflow, err := ResolveTransitionForTask(app, t, action, 0)
		if err != nil {
			return nil, err
		}

		ref := workflow.Start
		switch action {
		case "close":
			ref = workflow.Done
		case "reopen":
			ref = workflow.Open
		}
		group := groups[ref.TransitionKey]
		if group.ref.TransitionKey == "" {
			group.ref = ref
		}
		group.items = append(group.items, t)
		groups[ref.TransitionKey] = group
	}
	return groups, nil
}

func ResolveTransitionForTask(
	app *app.App, task *config.CacheItem, action string, attempt int,
) (*config.WorkflowMap, error) {

	workflow, err := app.LocalContext.GetWorkflowByType(task.Type)

	var focusedKey string
	if workflow != nil {
		switch action {
		case "close":
			focusedKey = workflow.Done.TransitionKey
		case "reopen":
			focusedKey = workflow.Open.TransitionKey
		default:
			focusedKey = workflow.Start.TransitionKey
		}
	}

	// Attempt to auto-discover the workflow once for the new task type.
	// In case the workflow transition cannot be found, error out to the user.
	if err != nil || focusedKey == "" {
		discovered, discoverErr := app.Client.DiscoverWorkflowForItem(app.Ctx, task)
		if discoverErr == nil {
			app.LocalContext.SetWorkflowForTaskType(app.Dir, task.Type, discovered)
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
