package cli

import (
	"fmt"

	"github.com/richo542/sneak/internal/config"
	"github.com/spf13/cobra"
)

func newShipCmd(app *App) *cobra.Command {
	var (
		all     bool
		message string
	)

	cmd := &cobra.Command{
		Use:   "ship",
		Short: "Close work item and issue git PR",
		Long: `Sets work items to 'closed' and creates a PR for the currently active branch.

Use '-m' to comment on the work item.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if app.LocalContext == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}

			return runShipCommand(
				app, args, all, message,
			)
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "close all active work items.")
	cmd.Flags().StringVarP(&message, "message", "m", "", "comment to add to the work item(s) to be started")

	return cmd
}

func runShipCommand(
	app *App, tasks []string,
	all bool, comment string,
) error {

	needsRefresh, err := CheckAndRefreshCache(app, false)
	if needsRefresh && err != nil {
		return err
	}

	cacheItems, err := ResolveTaskFocus(app, tasks, all)
	if err != nil {
		return err
	}

	activeItems := app.State.GetActiveTasksByCacheItems(cacheItems)
	if len(activeItems) != len(cacheItems) {
		fmt.Println(
			"cannot find all selected item(s) in active tasks. " +
				"'sneak ship' can only be used with managed tasks, consider 'sneak close' instead.",
		)
		return err
	}

	return nil
}

func processShipCommand(app *App, cacheItems []*config.CacheItem, comment string) error {

	if err := CloseCacheItems(app, cacheItems); err != nil {
		return err
	}

	// Steps
	// 1. Verify information about the cache item(s) are in the active task set with a branch.
	// 2. Resolve the active task branch
	// 3. Check if working tree is clean
	// 4. Prompt user with "continue anyway (y/n)"
	// 5. Create PR
	// 6. Add comment with PR Id or similar

	return nil
}

func resolveSharedBranch(items []*config.ActiveTask) (string, error) {
	branchName := ""
	for _, at := range items {
		if at.Branch != "" && branchName == "" {
			branchName = at.Branch
		}

		if branchName != at.Branch {
			return "", fmt.Errorf(
				"cannot identify shared branch name as tasks do not share the branch",
			)
		}
	}

	return branchName, nil
}
