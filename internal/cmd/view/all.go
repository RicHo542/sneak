package view

import (
	"errors"
	"fmt"
	"time"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/handlers"
	"github.com/richo542/sneak/internal/todos"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newAllCmd(app *app.App) *cobra.Command {
	var (
		refresh        bool
		todosFilter    bool
		providerFilter bool
	)

	cmd := &cobra.Command{
		Use:   "all",
		Short: "List all work items globally.",
		Long: `Displays all work items and todos accross all projects.

Uses a local cache (1hr TTL) for fast results.
Use --refresh to force a live fetch from the provider.
Use --todos to only show local todos, not provider items.
Use --remote to only show remote provider work items, not local todos.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAll(app, refresh, todosFilter, providerFilter)
		},
	}

	cmd.Flags().BoolVar(&refresh, "refresh", false, "force live fetch from provider")
	cmd.Flags().BoolVar(&todosFilter, "todos", false, "only show local todos")
	cmd.Flags().BoolVar(&providerFilter, "remote", false, "only show remote work items")

	return cmd
}

func runAll(
	instance *app.App, refresh bool,
	todoFilter bool, providerFilter bool,
) error {

	// If no flag is provided - List both.
	if !todoFilter && !providerFilter {
		todoFilter = true
		providerFilter = true
	}

	if instance.OutsideProjectScope() && providerFilter {
		return fmt.Errorf("cannot show provider items outside of project scope.")
	}

	var providerListErr error
	if providerFilter {
		providerListErr = runAllProviderItems(instance, refresh)
		if todoFilter {
			fmt.Println()
		}
	}

	var todoListErr error
	if todoFilter {
		todoListErr = runAllTodoItems(instance)
	}

	if todoListErr != nil || providerListErr != nil {
		return errors.Join(todoListErr, providerListErr)
	}

	return nil
}

func runAllTodoItems(instance *app.App) error {
	todoItems, err := instance.TodoStore.GetByStatus(
		instance.ProjectScope, []string{
			todos.TodoStatusOpen,
			todos.TodoStatusActive,
		},
	)
	if err != nil {
		return err
	}

	todos.SortForDisplay(todoItems)

	ui.PrintTableOfTodos(todoItems)
	ui.Printfln("%d todos in current context", len(todoItems))

	return nil
}

func runAllProviderItems(app *app.App, refresh bool) error {

	refreshRequired, err := handlers.CheckAndRefreshCache(app, refresh)
	if refreshRequired && err != nil {
		return err
	}

	items := app.State.Cache.Items
	if len(items) == 0 {
		fmt.Println("No work items found.")
		return nil
	}

	// Log out the age of this state for keep user informed
	age := time.Since(app.State.Cache.FetchedAt).Truncate(time.Second)

	ui.PrintTableOfProviderItems(items)
	ui.Printfln("%d work items (cached, fetched %s ago)", len(items), age)

	return nil
}
