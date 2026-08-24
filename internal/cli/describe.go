package cli

import (
	"fmt"

	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newDescribeCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe TASK_ID",
		Short: "Show details for a work item",
		Long: `Shows a detailed view of a single work item: name, description,
creation info, iteration/sprint, owner, and the most recent comments.

Always fetches live from the provider; never uses the local cache.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.LocalContext == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}
			return runDescribeCmd(app, args[0])
		},
	}

	return cmd
}

func runDescribeCmd(app *App, taskID string) error {
	detail, err := app.Client.DescribeWorkItem(app.Ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to describe %s: %w", taskID, err)
	}

	ui.PrintWorkItemDetail(detail)
	return nil
}
