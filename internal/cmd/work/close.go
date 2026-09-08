package work

import (
	"errors"
	"fmt"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/handlers"
	"github.com/richo542/sneak/internal/todos"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newCloseCmd(instance *app.App) *cobra.Command {
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

			if all && len(args) > 0 {
				return fmt.Errorf("'--all' can only be used without specifying task ids.")
			}

			if instance.OutsideProjectScope() && handlers.ContainsProviderKeys(args) {
				return fmt.Errorf("using 'close' outside project directories is limited to global todo items.")
			}

			return runCloseCmd(instance, args, all, message)
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

	providerItems, todoItems, err := handlers.ResolveTaskSelection(app, handlers.FocusClose, taskKeys, all)
	if err != nil {
		return err
	}

	var providerCloseErr error
	var todoCloseErr error

	if len(providerItems) > 0 {
		providerCloseErr = closeProviderItems(app, providerItems, comment)
	}

	if len(todoItems) > 0 {
		todoCloseErr = closeTodos(app, todoItems, comment)
	}

	if providerCloseErr != nil || todoCloseErr != nil {
		return errors.Join(providerCloseErr, todoCloseErr)
	}

	return nil
}

func closeProviderItems(app *app.App, cacheItems []*config.CacheItem, comment string) error {

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

func closeTodos(instance *app.App, todoItems []*todos.Todo, comment string) error {
	if err := instance.TodoStore.Close(todoItems); err != nil {
		return fmt.Errorf("failed to close todos: %w", err)
	}
	for _, item := range todoItems {
		ui.Printfln("closed todo: '%s'", item.Key)
	}
	if err := instance.TodoStore.AddNotes(todoItems, comment); err != nil {
		ui.Printfln("failed to add note: %v", err)
	}
	return nil
}
