package cli

import (
	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/cmd/adm"
	"github.com/richo542/sneak/internal/cmd/view"
	"github.com/richo542/sneak/internal/cmd/work"
	"github.com/spf13/cobra"
)

func NewRootCmd(info app.BuildInfo) *cobra.Command {
	appInst := &app.App{}

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
			return app.InitApp(cmd, appInst)
		},
		SilenceUsage: true,
	}
	root.CompletionOptions.DisableDefaultCmd = true

	adm.SetHelpCommand(root)
	adm.Register(appInst, info, root)
	view.Register(appInst, root)
	work.Register(appInst, root)

	return root
}
