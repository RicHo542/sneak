package work

import (
	"fmt"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/handlers"
	"github.com/spf13/cobra"
)

func newUnassignCmd(app *app.App) *cobra.Command {
	var (
		reopen  bool
		message string
	)

	cmd := &cobra.Command{
		Use:   "unassign TASK_ID...",
		Short: "Unassign active tasks",
		Long: `Clears the assignee of the given active tasks, removing them from your plate.

Use '--reopen' to also move the tasks back to the open (to-do) state.
Use '-m' to comment on the work items on unassignment.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.LocalContext == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}
			return runUnassignCmd(app, args, reopen, message)
		},
	}

	cmd.Flags().BoolVar(&reopen, "reopen", false, "move the tasks back to the open (to-do) state")
	cmd.Flags().StringVarP(&message, "message", "m", "", "comment to add to the work item(s) being unassigned")

	return cmd
}

func runUnassignCmd(
	app *app.App, taskKeys []string,
	reopen bool, comment string,
) error {

	refreshRequired, err := handlers.CheckAndRefreshCache(app, false)
	if refreshRequired && err != nil {
		return err
	}

	cacheItems, err := ResolveUnassignTaskFocus(app, taskKeys)
	if err != nil {
		return err
	}

	return processUnassignCmd(app, cacheItems, reopen, comment)
}

func processUnassignCmd(app *app.App, cacheItems []*config.CacheItem, reopen bool, comment string) error {
	if err := app.Client.UnassignWorkItems(app.Ctx, app.LocalContext, cacheItems); err != nil {
		return err
	}

	if reopen {
		if err := handlers.ReopenCacheItems(app, cacheItems); err != nil {
			return err
		}
	}

	handlers.CommentCacheItems(app, cacheItems, comment)

	app.State.RemoveActiveTasks(cacheItems)
	if err := app.SaveState(); err != nil {
		fmt.Println("Failed to save local state")
	}

	return nil
}

// ResolveUnassignTaskFocus resolves the given task keys to cache items,
// ensuring they are currently active tasks.
func ResolveUnassignTaskFocus(app *app.App, taskKeys []string) ([]*config.CacheItem, error) {
	if len(taskKeys) == 0 {
		return nil, fmt.Errorf("no task keys provided: unassign requires at least one active task")
	}

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
