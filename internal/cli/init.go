package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/richo542/sneak/internal/client"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

var abortedErr = errors.New("Aborted.")

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize sneak in the current directory",
		Long: `Link the current directory to a remote provider and work items.

Walks you through selecting a provider and configuring the project.
Creates .sneak/config.yaml which should be committed to git.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			return initialize(cmd.Context(), dir)
		},
	}

	return cmd
}

func initialize(ctx context.Context, dir string) error {
	reader := bufio.NewReader(os.Stdin)

	if err := checkAlreadyInitialized(reader, dir); err != nil {
		return err
	}

	provider, err := chooseProvider(reader)
	if err != nil {
		return err
	}

	cfg := &config.LocalContext{
		Remote: config.RemoteContext{
			Host: provider.Host,
			Type: provider.Type,
		},
	}

	switch provider.Type {
	case "jira":
		if err := initJiraConfig(reader, cfg); err != nil {
			return err
		}
	case "azure":
		if err := initAzureConfig(reader, cfg); err != nil {
			return err
		}
	}

	fmt.Println()
	itemsStr, err := ui.PromptLine(reader, "Parent item IDs (comma-separated, or leave empty)")
	if err != nil {
		return err
	}
	if itemsStr != "" {
		for id := range strings.SplitSeq(itemsStr, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				cfg.Bindings = append(cfg.Bindings, id)
			}
		}
	}

	fmt.Println()
	if err := discoverWorkflowTransitions(ctx, provider, cfg); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	fmt.Println()
	if err := config.StoreLocalContext(dir, cfg); err != nil {
		return err
	}

	fmt.Printf("Initialized .sneak/config.yaml for %s\n", provider.Host)
	return nil
}

func discoverWorkflowTransitions(
	ctx context.Context, provider *config.Provider, cfg *config.LocalContext,
) error {
	c, err := client.NewProviderClient(cfg, provider)
	if err != nil {
		return fmt.Errorf("could not discover workflow transitions: %w", err)
	}

	workflow, err := c.DiscoverWorkflow(ctx, cfg)
	if err != nil {
		return fmt.Errorf("could not discover workflow transitions: %w", err)
	}

	cfg.Transitions = workflow

	if len(workflow) == 0 {
		fmt.Println("No workflow transitions discovered.")
		return nil
	}

	types := make([]string, 0, len(workflow))
	for t := range workflow {
		types = append(types, t)
	}
	sort.Strings(types)

	fmt.Println("Discovered workflow transitions:")
	for _, t := range types {
		wm := workflow[t]
		fmt.Printf("  %-10s start=%-30s done=%s\n", t,
			transitionRefLabel(wm.Start), transitionRefLabel(wm.Done))
	}
	return nil
}

func transitionRefLabel(ref config.TransitionRef) string {
	if ref.DisplayName == "" {
		return "(unknown)"
	}
	if ref.TransitionKey != "" && ref.TransitionKey != ref.DisplayName {
		return fmt.Sprintf("%s (key %s)", ref.DisplayName, ref.TransitionKey)
	}
	return ref.DisplayName
}

func checkAlreadyInitialized(reader *bufio.Reader, dir string) error {
	configPath := filepath.Join(
		dir, config.LocalConfigDir, config.LocalConfigFile,
	)

	if _, err := os.Stat(configPath); err == nil {
		fmt.Print("Already initialized. Overwrite? [y/N] ")
		confirm, _ := ui.PromptLine(reader, "")
		if !strings.EqualFold(strings.TrimSpace(confirm), "y") {
			fmt.Println("Aborted.")
			return abortedErr
		}
	}
	return nil
}

func chooseProvider(reader *bufio.Reader) (*config.Provider, error) {
	providers, err := config.LoadProviders()
	if err != nil {
		return nil, fmt.Errorf("failed to load providers: %w", err)
	}

	if len(providers.Providers) == 0 {
		return nil, fmt.Errorf(
			"no providers configured. Run 'sneak config' first to set one up",
		)
	}

	var selected config.Provider

	// Automatically use the provider, if there is only one configured
	if len(providers.Providers) == 1 {
		for _, p := range providers.Providers {
			selected = p
		}
		fmt.Printf(
			"Using provider: %s (%s, %s)\n\n",
			selected.Alias, selected.Type, selected.Host,
		)

	} else {
		// List options with int-selection
		items := make([]ui.SelectItem, 0, len(providers.Providers))
		for _, p := range providers.Providers {
			items = append(items, ui.SelectItem{
				Key:   p.Alias,
				Label: fmt.Sprintf("%s  (%s, %s)", p.Alias, p.Type, p.Host),
			})
		}

		picked, err := ui.PromptSelect(reader, "Select provider", items)
		if err != nil {
			return nil, err
		}
		selected = providers.Providers[picked]
		fmt.Println()
	}

	return &selected, nil
}

func initJiraConfig(reader *bufio.Reader, cfg *config.LocalContext) error {
	project, err := ui.PromptLine(reader, "Project Name (e.g. PROJ)")
	if err != nil {
		return err
	}
	if strings.TrimSpace(project) == "" {
		return fmt.Errorf("project name is required")
	}
	cfg.Remote.Project = strings.TrimSpace(project)

	board, err := ui.PromptLine(reader, "Board ID (optional, press Enter to skip)")
	if err != nil {
		return err
	}
	if strings.TrimSpace(board) != "" {
		cfg.Remote.Board = strings.TrimSpace(board)
	}

	return nil
}

func initAzureConfig(reader *bufio.Reader, cfg *config.LocalContext) error {
	project, err := ui.PromptLine(reader, "Project name")
	if err != nil {
		return err
	}
	if strings.TrimSpace(project) == "" {
		return fmt.Errorf("project name is required")
	}
	cfg.Remote.Project = strings.TrimSpace(project)

	area, err := ui.PromptLine(reader, "Area path (optional, press Enter to skip)")
	if err != nil {
		return err
	}
	if strings.TrimSpace(area) != "" {
		cfg.Remote.AreaPath = strings.TrimSpace(area)
	}

	return nil
}
