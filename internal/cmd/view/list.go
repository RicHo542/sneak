package view

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/handlers"
	"github.com/richo542/sneak/internal/todos"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func Register(appInst *app.App, root *cobra.Command) {
	root.AddCommand(
		newListCmd(appInst),
		newDescribeCmd(appInst),
		newStatusCmd(appInst),
		newStandupCmd(),
		newOpenCmd(appInst),
		newAllCmd(appInst),
	)
}

func newListCmd(instance *app.App) *cobra.Command {
	var (
		refresh bool
		// typeFilter string
		todosFilter    bool
		providerFilter bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work items assigned to you",
		Long: `Displays work items under the configured bindings.

Uses a local cache (1hr TTL) for fast results.
Use --refresh to force a live fetch from the provider.
Use --todos to only show local todos, not provider items.
Use --remote to only show remote provider work items, not local todos.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(instance, refresh, todosFilter, providerFilter)
		},
	}

	cmd.Flags().BoolVar(&refresh, "refresh", false, "force live fetch from provider")
	cmd.Flags().BoolVar(&todosFilter, "todos", false, "only show local todos")
	cmd.Flags().BoolVar(&providerFilter, "remote", false, "only show remote work items")
	// cmd.Flags().StringVarP(&typeFilter, "types", "t", "", "filter by work item type (e.g. Story, Bug)")

	return cmd
}

func runList(
	instance *app.App, refresh bool,
	todoFilter bool, providerFilter bool,
) error {

	// Normalize the --todos / --remote filters against the current scope.
	todoFilter, providerFilter, err := validateFilterFlags(instance, todoFilter, providerFilter)
	if err != nil {
		return err
	}

	var providerListErr error
	if providerFilter {
		providerListErr = runListProviderItems(instance, refresh)
		if todoFilter {
			fmt.Println()
		}
	}

	var todoListErr error
	if todoFilter {
		todoListErr = runListTodoItems(instance)
	}

	if todoListErr != nil || providerListErr != nil {
		return errors.Join(todoListErr, providerListErr)
	}

	return nil
}

func runListTodoItems(instance *app.App) error {
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

func runListProviderItems(instance *app.App, refresh bool) error {

	refreshRequired, err := handlers.CheckAndRefreshCache(instance, refresh)
	if refreshRequired && err != nil {
		return err
	}

	state := instance.State
	// Shouldn't happen as setting bindings is mandatory during init,
	// however, keeping it as a saftey net.
	if len(state.Cache.Bindings) == 0 {
		return fmt.Errorf("No bindings set, please run 'sneak init' first for setup.")
	}

	items := state.Cache.Items
	if len(items) == 0 {
		fmt.Println("No work items found.")
		return nil
	}

	// Log out the age of this state for keep user informed
	age := time.Since(state.Cache.FetchedAt).Truncate(time.Second)

	ui.PrintTableOfProviderItems(items)
	ui.Printfln("%d work items (cached, fetched %s ago)", len(items), age)

	return nil
}

func validateFilterFlags(instance *app.App, todo, provider bool) (bool, bool, error) {
	if instance.InProjectScope() {
		// No filters → show both.
		if !todo && !provider {
			return true, true, nil
		}
		return todo, provider, nil
	}

	// Provider items require a project scope.
	if provider && !todo {
		return false, false, fmt.Errorf(
			"provider items cannot be displayed outside of project scope. " +
				"Use 'sneak all' for global overviews.",
		)
	}

	// Outside of a project only local todos can be shown; a requested or
	// defaulted provider view is dropped.
	switch {
	case provider:
		ui.Printfln(
			"provider items cannot be displayed outside of project scope. " +
				"Use 'sneak all' for global overviews.",
		)
	case !todo:
		ui.Printfln(
			"only showing local todo items. Use 'sneak all' for global overviews " +
				"of todos and provider items.",
		)
	}

	return true, false, nil
}

// parseTypes deprecated
func parseTypes(typeFilter string) []string {
	var types []string
	typesSplits := strings.SplitSeq(typeFilter, ",")
	for p := range typesSplits {
		if p = strings.TrimSpace(p); p != "" {
			types = append(types, p)
		}
	}

	return types
}

// filterByType deprecated
func filterByType(items []config.CacheItem, types []string) []config.CacheItem {
	var filtered []config.CacheItem
	for _, item := range items {
		for _, t := range types {
			if strings.EqualFold(item.Type, t) {
				filtered = append(filtered, item)
			}
		}
	}
	return filtered
}
