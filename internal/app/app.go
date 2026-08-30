package app

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

func InitApp(cmd *cobra.Command, app *App) error {
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

	c, err := client.NewProviderClient(localCtx, provider)
	if err != nil {
		return nil
	}
	app.Client = c

	state, err := config.LoadState(localCtx.ProjectID)
	if err != nil {
		return nil
	}
	app.State = state

	app.Ctx = cmd.Root().Context()
	return nil
}

// SaveState persists the current state, always stamping the project directory
// so it never reflects a stale value from a previous run.
func (app *App) SaveState() error {
	app.State.ProjectDir = app.Dir
	return config.SaveState(app.LocalContext.ProjectID, app.State)
}
