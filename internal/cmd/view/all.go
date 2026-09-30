package view

import (
	"errors"
	"fmt"
	"time"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/todos"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

type allFlags struct {
	local  bool
	remote bool
	active bool
	open   bool
	closed bool
}

func newAllCmd(app *app.App) *cobra.Command {
	var flags allFlags

	cmd := &cobra.Command{
		Use:   "all",
		Short: "List all work items globally.",
		Long: `Displays all work items and todos across all projects.

Reads from the local cache (1hr TTL) and shows the cache age per project.
Use --local to only show local todos, --remote to only show provider items.
Use --active, --open or --closed to filter by status.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAll(app, flags)
		},
	}

	cmd.Flags().BoolVar(&flags.local, "local", false, "only show local todos")
	cmd.Flags().BoolVar(&flags.remote, "remote", false, "only show remote work items")
	cmd.Flags().BoolVar(&flags.active, "active", false, "only show active items")
	cmd.Flags().BoolVar(&flags.open, "open", false, "only show open items")
	cmd.Flags().BoolVar(&flags.closed, "closed", false, "only show closed/done items")

	return cmd
}

func runAll(instance *app.App, flags allFlags) error {
	showTodos := !flags.remote || flags.local
	showProvider := !flags.local || flags.remote

	pathByScope := map[string]string{}
	states, err := config.DiscoverStates()
	if err != nil {
		return err
	}
	for _, st := range states {
		pathByScope[st.Key] = st.Dir
	}

	var errs []error

	var todoGroups []ui.TodoGroup
	if showTodos {
		todoGroups = collectAllTodos(instance, flags, pathByScope)
	}

	var providerGroups []ui.ProviderGroup
	if showProvider {
		groups, err := collectAllProviderItems(instance, states, flags)
		if err != nil {
			errs = append(errs, err)
		}
		providerGroups = groups
	}

	ui.PrintAllSummary(todoGroups, providerGroups)

	return errors.Join(errs...)
}

// collectAllTodos gathers open/active (or explicitly filtered) todos across
// every scope, grouped by their scope key and annotated with the project path.
// Scopes outside the state index (e.g. "global", or a todo-only project that
// never fetched) fall back to the scope key itself.
func collectAllTodos(instance *app.App, flags allFlags, pathByScope map[string]string) []ui.TodoGroup {
	wanted := wantedTodoStatuses(flags)
	buckets := instance.TodoStore.GetAllByStatus(wanted)

	groups := make([]ui.TodoGroup, 0, len(buckets))
	for _, b := range buckets {
		path, ok := pathByScope[b.Scope]
		if !ok {
			path = b.Scope
		}
		groups = append(groups, ui.TodoGroup{Project: b.Scope, Path: path, Items: b.Items})
	}
	return groups
}

// wantedTodoStatuses resolves the status filter for todos. With no status flag
// the default overview is open + active.
func wantedTodoStatuses(flags allFlags) []string {
	if !flags.open && !flags.active && !flags.closed {
		return []string{todos.TodoStatusOpen, todos.TodoStatusActive}
	}

	var wanted []string
	if flags.open {
		wanted = append(wanted, todos.TodoStatusOpen)
	}
	if flags.active {
		wanted = append(wanted, todos.TodoStatusActive)
	}
	if flags.closed {
		wanted = append(wanted, todos.TodoStatusClosed)
	}
	return wanted
}

// collectAllProviderItems gathers cached provider items across every project
// state discovered in the state dir. The current project's live (possibly
// refreshed) state is preferred when inside a project.
func collectAllProviderItems(instance *app.App, states []config.ActiveStates, flags allFlags) ([]ui.ProviderGroup, error) {
	var groups []ui.ProviderGroup
	for _, st := range states {
		var state *config.State
		switch {
		case instance.InProjectScope() && st.Key == instance.LocalContext.ProjectID:
			state = instance.State
		default:
			loaded, err := config.LoadState(st.Key)
			if err != nil {
				return nil, fmt.Errorf("failed to load state '%s': %w", st.Key, err)
			}
			state = loaded
		}

		items := filterProviderItems(state, flags)
		if len(items) == 0 {
			continue
		}

		groups = append(groups, ui.ProviderGroup{
			Project: st.Key,
			Path:    st.Dir,
			Items:   items,
			Age:     time.Since(state.Cache.FetchedAt).Truncate(time.Second),
		})
	}

	return groups, nil
}

func filterProviderItems(state *config.State, flags allFlags) []config.CacheItem {
	items := state.Cache.Items
	if len(items) == 0 {
		return nil
	}

	// Default overview: everything that is not done/closed.
	if !flags.open && !flags.active && !flags.closed {
		var filtered []config.CacheItem
		for _, item := range items {
			if todos.CoarseStatus(item.Status) != todos.TodoStatusClosed {
				filtered = append(filtered, item)
			}
		}
		return filtered
	}

	activeKeys := make(map[string]struct{}, len(state.ActiveTasks))
	for _, at := range state.ActiveTasks {
		activeKeys[at.Key] = struct{}{}
	}

	var filtered []config.CacheItem
	for _, item := range items {
		if flags.active {
			if _, ok := activeKeys[item.Key]; !ok {
				continue
			}
		}
		if flags.open && todos.CoarseStatus(item.Status) != todos.TodoStatusOpen {
			continue
		}
		if flags.closed && todos.CoarseStatus(item.Status) != todos.TodoStatusClosed {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered
}
