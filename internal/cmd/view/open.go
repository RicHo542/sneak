package view

import (
	"fmt"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newOpenCmd(app *app.App) *cobra.Command {
	var printURL bool

	cmd := &cobra.Command{
		Use:   "open TASK_ID",
		Short: "Open a work item in the browser",
		Long: `Opens the web page of a single work item in your default browser.

Fetches the item live from the provider to resolve its URL.
Use --print to only print the URL instead of opening the browser.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.LocalContext == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}
			return runOpenCmd(app, args[0], printURL)
		},
	}

	cmd.Flags().BoolVar(&printURL, "print", false, "print the URL instead of opening the browser")

	return cmd
}

func runOpenCmd(app *app.App, taskID string, printURL bool) error {
	detail, err := app.Client.DescribeWorkItem(app.Ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to resolve %s: %w", taskID, err)
	}

	if detail.URL == "" {
		return fmt.Errorf("no web URL available for %s", taskID)
	}

	if printURL {
		fmt.Println(detail.URL)
		return nil
	}

	ui.OpenBrowserOrPrint(detail.URL)
	return nil
}
