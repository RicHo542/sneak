package work

import (
	"fmt"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

type CreateOps struct {
	Title  string
	Type   string
	Parent string
	Labels []string
	Pin    bool
}

func newCreateCmd(app *app.App) *cobra.Command {
	var (
		remote   bool
		itemType string
		parent   string
		labels   []string
		pin      bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Todo or Work Item",
		Long: `Creates a new Todo or Work Item with the connected provider.

Use '--remote' to create a work item with the connected provider (Requires --parent and --type).
Without '--remote' a local Todo item is created.
Use '--labels' to assign labels to the created item.
`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) < 1 || args[0] == "" {
				return fmt.Errorf("no title supplied for item")
			}

			if remote {
				return runProviderCreateCommand(app, &CreateOps{
					Title:  args[0],
					Type:   itemType,
					Parent: parent,
					Labels: labels,
				})
			}

			return runLocalCreateCommand(app, &CreateOps{
				Title:  args[0],
				Type:   itemType,
				Parent: parent,
				Labels: labels,
				Pin:    pin,
			})
		},
	}

	cmd.Flags().BoolVarP(&remote, "remote", "r", false, "create a new work item with the remote provider")
	cmd.Flags().BoolVar(&pin, "pin", false, "pin task to the top of the list for priority tasks")
	cmd.Flags().StringVarP(&parent, "parent", "p", "", "parent to assign the new work item to for provider based items.")
	cmd.Flags().StringArrayVarP(&labels, "labels", "l", []string{}, "specify labels for the item")
	cmd.Flags().StringVarP(&itemType, "type", "t", "", "work item type to create for provider based items.")

	return cmd
}

func runProviderCreateCommand(
	instance *app.App, opts *CreateOps,
) error {
	return fmt.Errorf("provider task created currently not implemented. stay tuned.")
}

func runLocalCreateCommand(
	instance *app.App, opts *CreateOps,
) error {
	id, err := instance.TodoStore.Create(
		instance.ProjectScope,
		opts.Title,
		opts.Labels,
		opts.Pin,
	)
	if err != nil {
		return err
	}

	ui.Printfln("created todo: '%s': %s", id, opts.Title)
	return nil
}
