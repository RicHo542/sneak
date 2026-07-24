// internal/cli/root.go
package cli

import (
	"github.com/spf13/cobra"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// App holds shared dependencies every subcommand can use.
// This is what makes commands testable — inject fakes instead of
// real filesystem/network dependencies in tests.
type App struct {
	// Registry *registry.Registry
	// State    *state.Manager
	Verbose bool
}

func NewRootCmd(info BuildInfo) *cobra.Command {
	app := &App{}

	root := &cobra.Command{
		Use:     "sneak",
		Short:   "Manage CLI tools you need for your work",
		Version: info.Version,
		// PersistentPreRunE runs once, after flags are parsed, before any subcommand.
		// This is where we actually initialize the App now that we know flags like
		// --config or --verbose.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initApp(app, cmd)
		},
		SilenceUsage: true, // don't dump usage on every runtime error
	}

	root.PersistentFlags().StringVar(&configPath, "config", "", "path to sneak config dir (default ~/.sneak)")
	root.PersistentFlags().BoolVarP(&app.Verbose, "verbose", "v", false, "verbose output")

	root.AddCommand(
		newVersionCmd(info),
		newHelpCmd(),
		newConfigCmd(),
	)

	return root
}

var configPath string

func initApp(app *App, cmd *cobra.Command) error {
	/*
		reg, err := registry.Load(configPath)
		if err != nil {
			return err
		}
		app.Registry = reg


		st, err := state.NewManager(configPath)
		if err != nil {
			return err
		}
		app.State = st
	*/
	return nil
}
