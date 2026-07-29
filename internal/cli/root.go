// internal/cli/root.go
package cli

import (
	"os"

	"github.com/richo542/sneak/internal/client"
	"github.com/richo542/sneak/internal/config"
	"github.com/spf13/cobra"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

type App struct {
	Provider *config.Provider
	Client   client.ProviderClient
	Context  *config.Context
	State    *config.State
	Dir      string
	Verbose  bool
}

func NewRootCmd(info BuildInfo) *cobra.Command {
	app := &App{}

	root := &cobra.Command{
		Use:     "sneak",
		Short:   "A CLI to easily close work task without overhead",
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
		newInitCmd(),
		newListCmd(app),
	)

	return root
}

var configPath string

func initApp(app *App, cmd *cobra.Command) error {
	dir, err := os.Getwd()
	if err != nil {
		return nil
	}
	app.Dir = dir

	ctx, err := config.LoadContext(dir)
	if err != nil {
		return nil
	}
	app.Context = ctx

	provider, err := config.GetProviderByHost(ctx.Remote.Host)
	if err != nil {
		return nil
	}
	app.Provider = provider

	c, err := client.NewProviderClient(provider)
	if err != nil {
		return nil
	}
	app.Client = c

	state, err := config.LoadState(dir)
	if err != nil {
		return nil
	}
	app.State = state

	return nil
}
