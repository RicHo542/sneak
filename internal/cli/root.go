package cli

import (
	"context"
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
	Provider     *config.Provider
	Client       client.ProviderClient
	Ctx          context.Context
	LocalContext *config.LocalContext
	State        *config.State
	Dir          string
}

func NewRootCmd(info BuildInfo) *cobra.Command {
	app := &App{}

	root := &cobra.Command{
		Use:     "sneak",
		Short:   "A CLI to easily close work task without overhead",
		Version: info.Version,
		// Runs before any subcommand
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			switch cmd.Name() {
			case "version", "help", "config", "init":
				return nil
			}
			return initApp(cmd, app)
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

func initApp(cmd *cobra.Command, app *App) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	dir, err := config.FindProjectDir(currentDir)
	if err != nil {
		return err
	}
	app.Dir = dir

	localCtx, err := config.LoadContext(dir)
	if err != nil {
		return nil
	}
	app.LocalContext = localCtx

	provider, err := config.GetProviderByHost(localCtx.Remote.Host)
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

	app.Ctx = cmd.Root().Context()
	return nil
}
