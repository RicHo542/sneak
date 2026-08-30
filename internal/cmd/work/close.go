package work

import (
	"fmt"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/handlers"
	"github.com/spf13/cobra"
)

func newCloseCmd(app *app.App) *cobra.Command {
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

			if app.LocalContext == nil {
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
	app *app.App, taskKeys []string,
	all bool, comment string,
) error {

	refreshRequired, err := handlers.CheckAndRefreshCache(app, false)
	if refreshRequired && err != nil {
		return err
	}

	// resolve and forward
	cacheItems, err := handlers.ResolveCloseTaskFocus(app, taskKeys, all)
	if err != nil {
		return err
	}

	return processCloseCmd(app, cacheItems, comment)
}

func processCloseCmd(app *app.App, cacheItems []*config.CacheItem, comment string) error {

	if err := handlers.CloseCacheItems(app, cacheItems); err != nil {
		return err
	}

	handlers.CommentCacheItems(app, cacheItems, comment)

	app.State.RemoveActiveTasks(cacheItems)
	if err := app.SaveState(); err != nil {
		fmt.Println("Failed to save local state")
	}

	return nil
}
