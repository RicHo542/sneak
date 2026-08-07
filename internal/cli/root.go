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
}

func NewRootCmd(info BuildInfo) *cobra.Command {
	app := &App{}

	root := &cobra.Command{
		Use:     "sneak",
		Short:   "A CLI to easily close work task without overhead",
		Version: info.Version,
		// Runs before any subcommand
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initApp(app)
		},
		SilenceUsage: true,
	}

	root.AddCommand(
		newVersionCmd(info),
		newHelpCmd(),
		newConfigCmd(),
		newInitCmd(),
		newListCmd(app),
		newStartCmd(app),
		newCommentCmd(app),
		newCloseCmd(app),
	)

	return root
}

func initApp(app *App) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	dir, err := config.FindProjectDir(currentDir)
	if err != nil {
		return err
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
