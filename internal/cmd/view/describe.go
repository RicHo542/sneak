package view

import (
	"fmt"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/todos"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newDescribeCmd(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe TASK_ID",
		Short: "Show details for a work item or todo",
		Long: `Shows a detailed view of a single item: for provider work items the
name, description, creation info, iteration/sprint, owner, and the most recent
comments (always fetched live); for local todos the status, labels,
description, and notes.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDescribeCmd(app, args[0])
		},
	}

	return cmd
}

func runDescribeCmd(app *app.App, taskID string) error {
	if todos.HasTodoRef(taskID) {
		return runDescribeTodo(app, taskID)
	}

	if app.LocalContext == nil {
		return fmt.Errorf("not initialized: run 'sneak init' first")
	}

	detail, err := app.Client.DescribeWorkItem(app.Ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to describe %s: %w", taskID, err)
	}

	ui.PrintWorkItemDetail(detail)
	return nil
}

func runDescribeTodo(instance *app.App, ref string) error {
	item, err := instance.TodoStore.GetByRef(instance.ProjectScope, ref)
	if err != nil {
		return err
	}

	ui.PrintTodoDetail(item)
	return nil
}
