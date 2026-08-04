package cli

import (
	"fmt"
	"strings"

	"github.com/richo542/sneak/internal/config"
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
	invalidTasks, invalid := HasInvalidTasks(app, tasks)
	if invalid {
		for _, ivTask := range invalidTasks {
			fmt.Printf("task '%s' cannot be found.\n", ivTask)
		}
		return fmt.Errorf("Consider running 'sneak list --refresh' to update.")
	}

	return runNonInteractiveStartCommand(
		app, tasks, createBranch, comment,
	)
}

func runInteractiveStartCommand(app *App) error { return nil }

// transitionGroup groups work items by the transition key that moves them
// into their target state, so each group needs only one transition call.
type transitionGroup struct {
	ref   config.TransitionRef
	items []*config.CacheItem
}

func runNonInteractiveStartCommand(
	app *App, tasks []string, createBranch bool,
	comment string,
) error {

	cachedTasks, err := app.State.Cache.GetByKeyBatch(tasks)
	if err != nil {
		return fmt.Errorf(
			"unable to find tasks in cache. "+
				"Consider to run 'sneak list --refresh' to force a refresh.: %w",
			err,
		)
	}

	// Move tasks to in progress
	// Identify the transition target for each task by its type
	// Build groups with transitionKey -> work items
	groups, err := groupTasksByTransition(app, cachedTasks)
	if err != nil {
		return err
	}

	for _, group := range groups {
		if err := app.Client.TransitionWorkItems(app.Context, group.items, group.ref); err != nil {
			return fmt.Errorf("failed to move work items to in progress: %w", err)
		}

		// Keeps the cache updated and makes sure
		// active_tasks will also get the right state
		for _, item := range group.items {
			item.Status = group.ref.DisplayName
		}
	}

	// Create branch based on the task names
	branchName := ""
	if createBranch {
		createdBranchName, err := createBranchFromTasks(tasks)
		if err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
		branchName = createdBranchName
	}

	comment = strings.TrimSpace(comment)
	if comment != "" {
		if err := app.Client.AddCommentToWorkItems(
			nil, cachedTasks, comment,
		); err != nil {
			// Should this somehow fail?!
			fmt.Println("Failed to add comments to work items.")
		}
	}

	app.State.AddActiveTasks(cachedTasks, true, branchName)
	if err := config.SaveState(app.Dir, app.State); err != nil {
		fmt.Println("Failed to save tasks to local state")
	}

	return nil
}

func createBranchFromTasks(tasks []string) (string, error) {
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

func groupTasksByTransition(
	app *App, tasks []*config.CacheItem,
) (map[string]transitionGroup, error) {
	groups := make(map[string]transitionGroup)
	for _, t := range tasks {
		taskType := t.Type

		workflow, err := ResolveTransitionForType(app.Context, taskType)
		if err != nil {
			return nil, err
		}

		group := groups[workflow.Start.TransitionKey]
		if group.ref.TransitionKey == "" {
			group.ref = workflow.Start
		}
		group.items = append(group.items, t)
		groups[workflow.Start.TransitionKey] = group
	}
	return groups, nil
}

func ResolveTransitionForType(
	ctx *config.Context, taskType string,
) (*config.WorkflowMap, error) {

	workflow, err := ctx.GetWorkflowByType(taskType)
	if err != nil || workflow.Start.TransitionKey == "" {
		return nil, fmt.Errorf("unable to find workflow for task type '%s': %w", taskType, err)
		// Should this trigger the auto-resolve for transitions of this type? --> Yes
		// TODO Attempt to get Workflow config for this task type

	}

	return workflow, nil
}
