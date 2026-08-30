package work

import (
	"fmt"
	"strings"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/git"
	"github.com/richo542/sneak/internal/handlers"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func Register(appInst *app.App, root *cobra.Command) {
	root.AddCommand(
		newStartCmd(appInst),
		newCloseCmd(appInst),
		newShipCmd(appInst),
		newUnassignCmd(appInst),
		newCommentCmd(appInst),
	)
}

func newStartCmd(app *app.App) *cobra.Command {
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

			if app.LocalContext == nil {
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
	app *app.App, tasks []string,
	createBranch bool, comment string,
) error {

	// Check Cache expiry
	refreshRequired, err := handlers.CheckAndRefreshCache(app, false)
	if refreshRequired && err != nil {
		return err
	}

	cachedTasks, err := handlers.ResolveStartTaskFocus(app, tasks, false)
	if err != nil {
		return err
	}

	if err := processStartCommand(
		app, cachedTasks, createBranch, comment,
	); err != nil {
		return err
	}

	ui.Printfln("Started workitems: %s", strings.Join(tasks, ", "))
	return nil
}

func processStartCommand(
	app *app.App, cachedTasks []*config.CacheItem, createBranch bool,
	comment string,
) error {

	if err := handlers.StartCacheItems(app, cachedTasks); err != nil {
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

	// Do not fail if comments have not been added.
	// User is informed, but important transactions succeeded.
	handlers.CommentCacheItems(app, cachedTasks, comment)

	app.State.AddActiveTasks(cachedTasks, true, branchName)
	if err := app.SaveState(); err != nil {
		fmt.Println("Failed to save tasks to local state")
	}

	return nil
}

func createBranchFromTasks(tasks []*config.CacheItem) (string, error) {
	if !git.NewGitClient().IsRepo(".") {
		return "", fmt.Errorf("current context does not seem to be a git repository.")
	}

	branchName := git.BuildBranchName(tasks)
	gitClient := git.NewGitClient()

	if err := gitClient.CreateBranch(branchName); err != nil {
		return branchName, err
	}

	return branchName, nil
}
