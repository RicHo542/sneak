package work

import (
	"fmt"
	"strings"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/handlers"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newCommentCmd(app *app.App) *cobra.Command {
	var (
		comment string
	)

	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Comment on a given task",
		Long:  `Allows you to leave a comment on a work item.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if app.LocalContext == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}

			return runComment(app, args, comment)
		},
	}

	cmd.Flags().StringVarP(&comment, "message", "m", "", "message to add as a comment")
	cmd.MarkFlagRequired("message")

	return cmd
}

func runComment(app *app.App, tasks []string, comment string) error {

	refreshRequired, err := handlers.CheckAndRefreshCache(app, false)
	if refreshRequired && err != nil {
		return err
	}

	// Validate tasks names
	invalidTasks, invalid := handlers.HasInvalidTasks(app, tasks)
	if invalid {
		for _, ivTask := range invalidTasks {
			ui.Printfln("task '%s' cannot be found.", ivTask)
		}
		return fmt.Errorf("Consider running 'sneak list --refresh' to update.")
	}

	// Error can be ignored as it would raise in case an item
	// cannot be found - This is checked beforehand.
	cachedTasks, _ := app.State.Cache.GetByKeyBatch(tasks)

	comment = strings.TrimSpace(comment)
	if comment == "" {
		return fmt.Errorf("please provide a valid comment using '-m'.")
	}

	if err := app.Client.AddCommentToWorkItems(
		app.Ctx, app.LocalContext, cachedTasks, comment,
	); err != nil {
		ui.Printfln("Error: %w", err)
		return err
	}

	return nil
}
