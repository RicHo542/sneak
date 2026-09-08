package work

import (
	"errors"
	"fmt"
	"strings"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/git"
	"github.com/richo542/sneak/internal/handlers"
	"github.com/richo542/sneak/internal/todos"
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
		newCreateCmd(appInst),
	)
}

func newStartCmd(instance *app.App) *cobra.Command {
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

			if instance.OutsideProjectScope() && handlers.ContainsProviderKeys(args) {
				return fmt.Errorf("using 'start' outside project directories is limited to global todo items.")
			}

			return runStartCommand(instance, args, createBranch, message)
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

	providerItems, todoItems, err := handlers.ResolveTaskSelection(
		app, handlers.FocusStart, tasks, false,
	)
	if err != nil {
		return err
	}

	if createBranch && len(providerItems) == 0 {
		return fmt.Errorf("'--branch' can only be used with provider work items.")
	}

	var providerStartErr error
	var todoStartErr error

	if len(providerItems) > 0 {
		providerStartErr = processProviderStart(
			app, providerItems, createBranch, comment,
		)
	}

	if len(todoItems) > 0 {
		todoStartErr = startTodos(app, todoItems, comment)
	}

	if providerStartErr != nil || todoStartErr != nil {
		return errors.Join(providerStartErr, todoStartErr)
	}

	ui.Printfln("Started workitems: %s", strings.Join(tasks, ", "))
	return nil
}

func startTodos(instance *app.App, todoItems []*todos.Todo, comment string) error {
	if err := instance.TodoStore.Start(todoItems); err != nil {
		return fmt.Errorf("failed to start todos: %w", err)
	}
	for _, item := range todoItems {
		ui.Printfln("started todo: '%s'", item.Key)
	}
	if err := instance.TodoStore.AddNotes(todoItems, comment); err != nil {
		ui.Printfln("failed to add note: %v", err)
	}
	return nil
}

func processProviderStart(
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
