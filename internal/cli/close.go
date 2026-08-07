package cli

import (
	"fmt"
	"strings"

	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newCloseCmd(app *App) *cobra.Command {
	var (
		all     bool
		message string
	)

	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close assigned tasks",
		Long: `Closes work items, setting them to status 'done'.

Use '-a' to close all active tasks, managed by sneak.
Use '-m' to comment on the work items.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if app.Context == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}

			return runCloseCmd(app, args, all, message)
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "close all active, managed tasks")
	cmd.Flags().StringVarP(&message, "message", "m", "", "comment to add to the work item(s) to be closed")

	return cmd
}

func runCloseCmd(
	app *App, taskKeys []string,
	all bool, comment string,
) error {

	refreshRequired, err := CheckAndRefreshCache(app, false)
	if refreshRequired && err != nil {
		return err
	}

	// resolve and forward
	cacheItems, err := resolveTaskFocus(app, taskKeys, all)
	if err != nil {
		return err
	}

	return processCloseCmd(app, cacheItems, comment)
}

func resolveTaskFocus(app *App, taskKeys []string, all bool) ([]*config.CacheItem, error) {
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

func processCloseCmd(app *App, cacheItems []*config.CacheItem, comment string) error {

	if err := TransitionCacheItems(app, cacheItems, "close"); err != nil {
		return err
	}

	comment = strings.TrimSpace(comment)
	if comment != "" {
		_ = app.Client.AddCommentToWorkItems(
			app.Context, cacheItems, comment,
		)
	}

	app.State.RemoveActiveTasks(cacheItems)
	if err := config.SaveState(app.Dir, app.State); err != nil {
		fmt.Println("Failed to save local state")
	}

	return nil
}
