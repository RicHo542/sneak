package cli

import (
	"fmt"
	"time"

	"github.com/richo542/sneak/internal/client/objects"
	"github.com/richo542/sneak/internal/config"
)

func CheckAndRefreshCache(app *App, forceRefresh bool) (bool, error) {
	state := app.State
	bindingsChanged := !state.Cache.MatchesBindings(app.LocalContext.Bindings)
	needsFetch := forceRefresh || bindingsChanged || !state.Cache.IsFresh()

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

	items, err := app.Client.ListWorkItems(app.Ctx, app.LocalContext, objects.ListOptions{
		Bindings: app.LocalContext.Bindings,
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
		Bindings:  app.LocalContext.Bindings,
	}

	reconcileActiveTasks(state, cacheItems)

	if err := config.SaveState(app.Dir, state); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	return nil
}

func reconcileActiveTasks(state *config.State, cacheItems []config.CacheItem) {
	statusByKey := make(map[string]string, len(cacheItems))
	for _, item := range cacheItems {
		statusByKey[item.Key] = item.Status
	}

	// Sync statuses for items still in the cache.
	for i := range state.ActiveTasks {
		at := &state.ActiveTasks[i]
		if status, ok := statusByKey[at.Key]; ok && status != at.Status {
			at.Status = status
		}
	}

	// Remove active tasks no longer present in the cache.
	kept := state.ActiveTasks[:0]
	for _, at := range state.ActiveTasks {
		if _, ok := statusByKey[at.Key]; ok {
			kept = append(kept, at)
		}
	}
	state.ActiveTasks = kept
}
