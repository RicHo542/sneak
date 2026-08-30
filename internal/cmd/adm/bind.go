package adm

import (
	"fmt"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/handlers"
	"github.com/spf13/cobra"
)

func newBindCmd(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "Manage parent item bindings",
		Long: `Manage the parent item IDs this project is linked to.

With no sub-command, prints the current bindings.
Sub-commands:
  list   list the current parent item IDs
  add    add one or more parent item IDs
  remove remove one or more parent item IDs
  set    replace the full binding list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.LocalContext == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}
			runBindList(app)
			return nil
		},
	}

	cmd.AddCommand(
		newBindListCmd(app),
		newBindAddCmd(app),
		newBindRemoveCmd(app),
		newBindSetCmd(app),
	)

	return cmd
}

func newBindListCmd(app *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List parent item bindings",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.LocalContext == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}
			runBindList(app)
			return nil
		},
	}
}

func newBindAddCmd(app *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "add <ID>...",
		Short: "Add parent item bindings",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.LocalContext == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}
			return mutateBindings(app, appendBindings, args)
		},
	}
}

func newBindRemoveCmd(app *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <ID>...",
		Short: "Remove parent item bindings",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.LocalContext == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}
			return mutateBindings(app, removeBindings, args)
		},
	}
}

func newBindSetCmd(app *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "set <ID>...",
		Short: "Replace all parent item bindings",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.LocalContext == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}
			return mutateBindings(app, nil, args)
		},
	}
}

func runBindList(app *app.App) {
	if len(app.LocalContext.Bindings) == 0 {
		fmt.Println("No bindings set.")
		return
	}
	fmt.Println("Bindings:")
	for _, b := range app.LocalContext.Bindings {
		fmt.Printf("  %s\n", b)
	}
}

// mutateBindings transforms the current bindings (replace with a nil mutator),
// persists the local context, and forces a cache refresh so the new scope is
// reflected in subsequent commands.
func mutateBindings(app *app.App, mutate func(values []string, args []string) []string, args []string) error {
	if mutate != nil {
		app.LocalContext.Bindings = mutate(app.LocalContext.Bindings, args)
	} else {
		app.LocalContext.Bindings = append([]string(nil), args...)
	}

	// Deduplicate without reordering.
	app.LocalContext.Bindings = uniqueStrings(app.LocalContext.Bindings)

	if err := config.StoreLocalContext(app.Dir, app.LocalContext); err != nil {
		return fmt.Errorf("failed to persist bindings: %w", err)
	}

	if err := handlers.RefreshCache(app); err != nil {
		return fmt.Errorf("failed to refresh cache after binding change: %w", err)
	}

	runBindList(app)
	return nil
}

func appendBindings(bindings []string, args []string) []string {
	return append(bindings, args...)
}

func removeBindings(bindings []string, args []string) []string {
	remove := make(map[string]bool, len(args))
	for _, a := range args {
		remove[a] = true
	}
	var out []string
	for _, b := range bindings {
		if !remove[b] {
			out = append(out, b)
		}
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
