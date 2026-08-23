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
	fmt.Println(strings.Repeat("-", 110))

	for _, p := range cfg.Providers {
		status := "not tested"

		providerClient, err := client.NewProviderClient(nil, &p)
		if err != nil {
			status = fmt.Sprintf("failed: %s", err)
		}

		if err := providerClient.TestConnection(); err != nil {
			status = fmt.Sprintf("error: %s", err)
		} else {
			status = "ok"
		}

		h := p.Host
		if len(h) > 40 {
			h = h[:37] + "..."
		}

		u := p.UserHandle
		if len(u) > 24 {
			u = u[:21] + "..."
		}

		fmt.Printf("%-16s  %-8s  %-40s  %-24s  %s\n", p.Alias, p.Type, h, u, status)
	}

	return nil
}

func runConfigSetup() error {
	reader := bufio.NewReader(os.Stdin)

	ui.PrintBanner()

	provider, err := promptProviderInfo(reader)
	if err != nil {
		return err
	}

	if provider.Type == "azure" {
		if err := addAzureSpecificProviderInfo(reader, provider); err != nil {
			return err
		}
	}

	fmt.Println()
	fmt.Print("Testing connection... ")

	providerClient, err := client.NewProviderClient(nil, provider)
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

	if userInfo, err := providerClient.GetUserIdent(); err == nil {
		provider.UserHandle = userInfo.UserHandle
		provider.UserDisplayName = userInfo.DisplayName

		fmt.Printf("Authenticated as: %s (%s)\n", provider.UserDisplayName, provider.UserHandle)
	} else {
		fmt.Printf("Warning: could not resolve current user: %v\n", err)
	}

	if err := persistProvider(provider); err != nil {
		return err
	}

	return nil
}

func promptProviderInfo(reader *bufio.Reader) (*config.Provider, error) {
	providerType, err := promptProviderType(reader)
	if err != nil {
		return nil, err
	}

	alias, err := ui.PromptLine(reader, "Alias (short name, e.g. 'work-jira')")
	if err != nil || alias == "" {
		return nil, fmt.Errorf("alias invalid or empty: %v", err)
	}

	host, err := ui.PromptLine(reader, "Host URL (e.g. https://mycompany.atlassian.net)")
	if err != nil || host == "" {
		return nil, fmt.Errorf("host invalid or empty: %v", err)
	}

	username, err := ui.PromptLine(reader, "Username / email")
	if err != nil || username == "" {
		return nil, fmt.Errorf("username invalid or empty: %v", err)
	}

	fmt.Print("Password / PAT: ")
	token, err := ui.ReadSecret()
	if err != nil || token == "" {
		return nil, fmt.Errorf("token invalid or empty: %v", err)
	}

	provider := config.Provider{
		Alias:    strings.TrimSpace(alias),
		Type:     providerType,
		Host:     strings.TrimSpace(host),
		Username: strings.TrimSpace(username),
		Token:    token,
	}

	return &provider, nil
}

func promptProviderType(reader *bufio.Reader) (string, error) {

	options := make([]ui.SelectItem, 2)
	options[0] = ui.SelectItem{Key: "jira", Label: "Jira"}
	options[1] = ui.SelectItem{Key: "azure", Label: "Azure"}

	selection, err := ui.PromptSelect(reader, "Provider type", options)
	if err != nil {
		return "", err
	}
	return selection, nil
}

func addAzureSpecificProviderInfo(reader *bufio.Reader, prov *config.Provider) error {
	org, err := ui.PromptLine(reader, "Organization name")
	if err != nil {
		return err
	}
	if strings.TrimSpace(org) == "" {
		return fmt.Errorf("organization name is required")
	}
	prov.Organization = strings.TrimSpace(org)
	return nil
}

func persistProvider(provider *config.Provider) error {
	cfg, err := config.LoadProviders()
	if err != nil {
		return err
	}

	cfg.Providers[provider.Alias] = *provider

	if err := config.SaveProviders(cfg); err != nil {
		return err
	}

	path, _ := config.SneakConfigDir()
	fmt.Printf("\nProvider '%s' saved → %s/providers.toml\n", provider.Alias, path)
	return nil
}
