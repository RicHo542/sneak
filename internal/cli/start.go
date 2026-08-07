package cli

import (
	"fmt"
	"strings"

	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/git"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

// transitionGroup groups work items by the transition key that moves them
// into their target state, so each group needs only one transition call.
type transitionGroup struct {
	ref   config.TransitionRef
	items []*config.CacheItem
}

func newStartCmd(app *App) *cobra.Command {
	var (
		createBranch bool
		message      string
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Assign and start work item",
		Long: `Assigns and sets a work item to 'in progress'.

Use '-b' to also create a new feature branch to directly create.
Use '-m' to comment on the work item.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if app.Context == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}

			return runStartCommand(app, args, createBranch, message)
		},
	}

	cmd.Flags().BoolVarP(&createBranch, "branch", "b", false, "create a feature branch in current git.")
	cmd.Flags().StringVarP(&message, "message", "m", "", "comment to add to the work item(s) to be started")

	return cmd
}

func runStartCommand(
	app *App, tasks []string,
	createBranch bool, comment string,
) error {

	// Check Cache expiry
	refreshRequired, err := CheckAndRefreshCache(app, false)
	if refreshRequired && err != nil {
		return err
	}

	// Prompt user with interactive selection if no key was provided.
	if len(tasks) < 1 {
		tasks, err = ui.InteractiveSelectItem(app.State.Cache.Items)
		if err != nil {
			return err
		}
	}

	cachedTasks, err := app.State.Cache.GetByKeyBatch(tasks)
	if err != nil {
		return fmt.Errorf(
			"unable to find tasks in cache. "+
				"Consider to run 'sneak list --refresh' to force a refresh.: %w",
			err,
		)
	}

	return processStartCommand(
		app, cachedTasks, createBranch, comment,
	)
}

func processStartCommand(
	app *App, cachedTasks []*config.CacheItem, createBranch bool,
	comment string,
) error {

	if err := TransitionCacheItems(app, cachedTasks, "start"); err != nil {
		return err
	}

	// Create branch based on the task names
	branchName := ""
	if createBranch {
		createdBranchName, err := createBranchFromTasks(cachedTasks)
		if err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
		branchName = createdBranchName
	}

	comment = strings.TrimSpace(comment)
	if comment != "" {
		_ = app.Client.AddCommentToWorkItems(
			app.Context, cachedTasks, comment,
		)
	}

	app.State.AddActiveTasks(cachedTasks, true, branchName)
	if err := config.SaveState(app.Dir, app.State); err != nil {
		fmt.Println("Failed to save tasks to local state")
	}

	return nil
}

func createBranchFromTasks(tasks []*config.CacheItem) (string, error) {
	if !git.NewGitClient().IsRepo() {
		return "", fmt.Errorf("current context does not seem to be a git repository.")
	}

	branchName := git.BuildBranchName(tasks)
	gitClient := git.NewGitClient()

	if err := gitClient.CreateBranch(branchName); err != nil {
		return branchName, err
	}

	return branchName, nil
}
