package work

import (
	"fmt"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/todos"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newNoteCmd(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "note",
		Short: "Edit the description of a local todo",
		Long: `Opens the editor to write or update the description of a local todo.

Provider work items are not yet supported for description editing.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNote(app, args[0])
		},
	}

	return cmd
}

func runNote(instance *app.App, ref string) error {
	if !todos.HasTodoRef(ref) {
		return fmt.Errorf(
			"provider description editing is not yet supported. "+
				"Use 'sneak describe %s' to view the current description.", ref,
		)
	}

	item, err := instance.TodoStore.GetByRef(instance.ProjectScope, ref)
	if err != nil {
		return err
	}

	newDoc, err := ui.EditText(item.Description, "")
	if err != nil {
		return err
	}

	if newDoc == item.Description {
		fmt.Println("no changes.")
		return nil
	}

	if err := instance.TodoStore.UpdateDescription(
		instance.ProjectScope, ref, newDoc,
	); err != nil {
		return fmt.Errorf("failed to save description: %w", err)
	}

	ui.Printfln("updated description of '%s'", ref)
	return nil
}
