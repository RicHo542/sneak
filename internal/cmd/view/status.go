package view

import (
	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newStatusCmd(app *app.App) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Lists the currently active tasks",
		Long: `List all tasks that are currently set to active and are managed by sneak.

Useful to get a quick glance at what to close and what was started last time you picked up your codebase.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusCmd(app)
		},
	}

	return cmd
}

func runStatusCmd(app *app.App) error {
	ui.PrintActiveTaskTable(app.State.ActiveTasks)
	ui.Printfln("%d active items", len(app.State.ActiveTasks))
	return nil
}
