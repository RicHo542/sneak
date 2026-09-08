package work

import (
	"errors"
	"fmt"
	"strings"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/handlers"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newCommentCmd(app *app.App) *cobra.Command {
	var (
		message string
	)

	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Comment on a given task",
		Long:  `Allows you to leave a comment on a provider work item or a note on a local todo.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runComment(app, args, message)
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "message to add as a comment")
	cmd.MarkFlagRequired("message")

	return cmd
}

func runComment(instance *app.App, tasks []string, comment string) error {

	comment = strings.TrimSpace(comment)
	if comment == "" {
		return fmt.Errorf("please provide a valid comment using '-m'.")
	}

	if instance.OutsideProjectScope() && handlers.ContainsProviderKeys(tasks) {
		return fmt.Errorf("using 'comment' outside project directories is limited to global todo items.")
	}

	providerItems, todoItems, err := handlers.ResolveCommentTargets(instance, tasks)
	if err != nil {
		return err
	}

	var errs []error
	if len(providerItems) > 0 {
		if err := instance.Client.AddCommentToWorkItems(
			instance.Ctx, instance.LocalContext, providerItems, comment,
		); err != nil {
			errs = append(errs, err)
		}
	}

	if len(todoItems) > 0 {
		if err := instance.TodoStore.AddNotes(todoItems, comment); err != nil {
			errs = append(errs, fmt.Errorf("failed to add note: %w", err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	ui.Printfln("commented on %d item(s)", len(providerItems)+len(todoItems))
	return nil
}
