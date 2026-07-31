package cli

import (
	"fmt"

	"github.com/richo542/sneak/internal/git"
	"github.com/spf13/cobra"
)

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

	if len(tasks) < 1 {
		return runInteractiveStartCommand(app)
	}

	// Validate tasks names
	invalidTasks, hasInvalid := HasInvalidTasks(app, tasks)
	if hasInvalid {
		for _, ivTask := range invalidTasks {
			fmt.Printf("task '%s' cannot be found.", ivTask)
		}
		return fmt.Errorf("some tasks in selection are invalid.")
	}

	return runNonInteractiveStartCommand(app, tasks, createBranch, comment)
}

func runInteractiveStartCommand(app *App) error { return nil }

func runNonInteractiveStartCommand(app *App, tasks []string, createBranch bool, comment string) error {
	if createBranch {
		if err := createBranchFromTasks(tasks); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
	}

	return nil
}

func createBranchFromTasks(tasks []string) error {
	if !git.NewGitClient().IsRepo() {
		return fmt.Errorf("current context does not seem to be a git repository.")
	}

	branchName := git.BuildBranchName(tasks)
	gitClient := git.NewGitClient()

	if err := gitClient.CreateBranch(branchName); err != nil {
		return err
	}

	return nil
}
