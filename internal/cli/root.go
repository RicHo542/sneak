package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCmd(info BuildInfo) *cobra.Command {
	app := &App{}

	root := &cobra.Command{
		Use:     "sneak",
		Short:   "A CLI to easily close work task without overhead",
		Version: info.Version,
		// Runs before any subcommand
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			switch cmd.Name() {
			case "version", "help", "config", "init", "sup":
				return nil
			}
			return InitApp(cmd, app)
		},
		SilenceUsage: true,
	}
	root.CompletionOptions.DisableDefaultCmd = true

	root.SetHelpCommand(newHelpCmd())
	root.AddCommand(
		newVersionCmd(info),
		newConfigCmd(),
		newInitCmd(),
		newListCmd(app),
		newStartCmd(app),
		newCommentCmd(app),
		newCloseCmd(app),
		newStatusCmd(app),
		newDescribeCmd(app),
		newOpenCmd(app),
		newBindCmd(app),
		newUnassignCmd(app),
		newStandupCmd(),
	)

	return root
}
