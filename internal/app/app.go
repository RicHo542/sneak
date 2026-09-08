package app

import (
	"context"
	"fmt"
	"os"

	"github.com/richo542/sneak/internal/client"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/todos"
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
	TodoStore    *todos.TodoStore
	Dir          string
	ProjectScope string
}

const GlobalProjectContext = "global"

// SaveState persists the current state, always stamping the project directory
// so it never reflects a stale value from a previous run.
func (app *App) SaveState() error {
	app.State.ProjectDir = app.Dir
	return config.SaveState(app.LocalContext.ProjectID, app.State)
}

func (app *App) InProjectScope() bool {
	return app.ProjectScope != GlobalProjectContext
}

func (app *App) OutsideProjectScope() bool {
	return app.ProjectScope == GlobalProjectContext
}

func InitApp(cmd *cobra.Command, app *App) error {
	dir, err := resolveProjectContext()

	if dir != "" && err == nil {
		app.Dir = dir
		loadProviderBoundContexts(app)
		app.ProjectScope = app.LocalContext.ProjectID
	} else {
		app.ProjectScope = GlobalProjectContext
	}

	loadTodoContexts(app)

	app.Ctx = cmd.Root().Context()
	return nil
}

func resolveProjectContext() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir, err := config.FindProjectDir(currentDir)
	if err != nil {
		return "", err
	}

	return dir, nil
}

func loadProviderBoundContexts(app *App) error {

	localCtx, err := config.LoadContext(app.Dir)
	if err != nil {
		return fmt.Errorf("unable to read .sneak/config.yaml: %v", err)
	}
	app.LocalContext = localCtx

	provider, err := config.GetProviderByHost(localCtx.Remote.Host)
	if err != nil {
		return fmt.Errorf("unable to identify provider: %s", localCtx.Remote.Host)
	}
	app.Provider = provider

	c, err := client.NewProviderClient(localCtx, provider)
	if err != nil {
		return fmt.Errorf("unable to connect provider: %s", provider.Host)
	}
	app.Client = c

	state, err := config.LoadState(localCtx.ProjectID)
	if err != nil {
		return fmt.Errorf("unable to load state '%s': %v", localCtx.ProjectID, err)
	}
	app.State = state

	return nil
}

func loadTodoContexts(app *App) error {
	todoStore, err := todos.Load()
	if err != nil {
		return fmt.Errorf("unable to read todo state: %v", err)
	}
	app.TodoStore = todoStore
	return nil
}
