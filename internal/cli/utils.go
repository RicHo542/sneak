package cli

import (
	"fmt"
	"time"

	"github.com/richo542/sneak/internal/client"
	"github.com/richo542/sneak/internal/config"
)

func RefreshCache(app *App) error {
	fmt.Println("Fetching work items...")
	state := app.State

	items, err := app.Client.ListWorkItems(app.Context, client.ListOptions{
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
