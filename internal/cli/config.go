package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/richo542/sneak/internal/client"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	var list bool

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage provider connections (Azure, Jira)",
		Long: `Configure and test connections to issue tracking providers.

Without flags, runs an interactive setup wizard to add or update a provider.
Use --list to show all configured providers and their connection status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				return runConfigList()
			}
			return runConfigSetup()
		},
	}

	cmd.Flags().BoolVar(&list, "list", false, "list configured providers and test their connections")

	return cmd
}

func runConfigList() error {
	cfg, err := config.LoadProviders()
	if err != nil {
		return err
	}

	if len(cfg.Providers) == 0 {
		fmt.Println("No providers configured. Run 'sneak config' to add one.")
		return nil
	}

	fmt.Printf("%-16s  %-8s  %-40s  %-24s  %s\n", "ALIAS", "TYPE", "HOST", "USER", "STATUS")
	fmt.Println(strings.Repeat("-", 96))

	for _, p := range cfg.Providers {
		status := "not tested"

		providerClient, err := client.NewProviderClient(&p)
		if err != nil {
			status = fmt.Sprintf("failed: %s", err)
		}

		if err := providerClient.TestConnection(); err != nil {
			status = fmt.Sprintf("error: %s", err)
		} else {
			status = "ok"
		}
		fmt.Printf("%-16s  %-8s  %-40s  %-24s  %s\n", p.Alias, p.Type, p.Host, p.UserHandle, status)
	}

	return nil
}

func runConfigSetup() error {
	reader := bufio.NewReader(os.Stdin)

	ui.PrintBanner()

	providerType, err := ui.PromptChoice(reader, "Provider type", []string{"jira", "azure"})
	if err != nil {
		return err
	}

	alias, err := ui.PromptLine(reader, "Alias (short name, e.g. 'work-jira')")
	if err != nil {
		return err
	}
	if strings.TrimSpace(alias) == "" {
		return fmt.Errorf("alias cannot be empty")
	}

	host, err := ui.PromptLine(reader, "Host URL (e.g. https://mycompany.atlassian.net)")
	if err != nil {
		return err
	}
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("host cannot be empty")
	}

	username, err := ui.PromptLine(reader, "Username / email")
	if err != nil {
		return err
	}
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username cannot be empty")
	}

	fmt.Print("Password / PAT: ")
	token, err := ui.ReadSecret()
	if err != nil {
		return err
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("token cannot be empty")
	}

	provider := config.Provider{
		Alias:    strings.TrimSpace(alias),
		Type:     providerType,
		Host:     strings.TrimSpace(host),
		Username: strings.TrimSpace(username),
		Token:    token,
	}

	fmt.Println()
	fmt.Print("Testing connection... ")

	providerClient, err := client.NewProviderClient(&provider)
	if err != nil {
		return err
	}

	if err := providerClient.TestConnection(); err != nil {
		fmt.Printf("failed\n  → %s\n\n", err)
		fmt.Print("Save provider anyway? [y/N] ")
		confirm, _ := ui.PromptLine(reader, "")
		if !strings.EqualFold(strings.TrimSpace(confirm), "y") {
			return fmt.Errorf("aborted — provider not saved")
		}
	} else {
		fmt.Println("ok")
	}

	if userHandle, err := providerClient.GetUserIdent(); err == nil {
		provider.UserHandle = userHandle
		fmt.Printf("Authenticated as: %s\n", userHandle)
	} else {
		fmt.Printf("Warning: could not resolve current user: %v\n", err)
	}

	if err := persistProvider(provider); err != nil {
		return err
	}

	return nil
}

func persistProvider(provider config.Provider) error {
	cfg, err := config.LoadProviders()
	if err != nil {
		return err
	}

	cfg.Providers[provider.Alias] = provider

	if err := config.SaveProviders(cfg); err != nil {
		return err
	}

	path, _ := config.SneakConfigDir()
	fmt.Printf("\nProvider '%s' saved → %s/providers.toml\n", provider.Alias, path)
	return nil
}
