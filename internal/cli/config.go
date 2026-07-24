package cli

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/richo542/sneak/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

// --- List ---

func runConfigList() error {
	cfg, err := config.LoadProviders()
	if err != nil {
		return err
	}

	if len(cfg.Providers) == 0 {
		fmt.Println("No providers configured. Run 'sneak config' to add one.")
		return nil
	}

	fmt.Printf("%-16s  %-8s  %-40s  %s\n", "ALIAS", "TYPE", "HOST", "STATUS")
	fmt.Println(strings.Repeat("-", 80))

	for _, p := range cfg.Providers {
		status := "not tested"
		if err := testConnection(p); err != nil {
			status = fmt.Sprintf("error: %s", err)
		} else {
			status = "ok"
		}
		fmt.Printf("%-16s  %-8s  %-40s  %s\n", p.Alias, p.Type, p.Host, status)
	}

	return nil
}

// --- Interactive Setup ---

func runConfigSetup() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║       sneak provider setup           ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()

	providerType, err := promptChoice(reader, "Provider type", []string{"jira", "azure"})
	if err != nil {
		return err
	}

	alias, err := promptLine(reader, "Alias (short name, e.g. 'work-jira')")
	if err != nil {
		return err
	}
	if strings.TrimSpace(alias) == "" {
		return fmt.Errorf("alias cannot be empty")
	}

	host, err := promptLine(reader, "Host URL (e.g. https://mycompany.atlassian.net)")
	if err != nil {
		return err
	}
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("host cannot be empty")
	}

	username, err := promptLine(reader, "Username / email")
	if err != nil {
		return err
	}
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username cannot be empty")
	}

	fmt.Print("Password / PAT: ")
	token, err := readSecret()
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
	if err := testConnection(provider); err != nil {
		fmt.Printf("failed\n  → %s\n\n", err)
		fmt.Print("Save provider anyway? [y/N] ")
		confirm, _ := promptLine(reader, "")
		if !strings.EqualFold(strings.TrimSpace(confirm), "y") {
			return fmt.Errorf("aborted — provider not saved")
		}
	} else {
		fmt.Println("ok")
	}

	cfg, err := config.LoadProviders()
	if err != nil {
		return err
	}

	updated := false
	for i, p := range cfg.Providers {
		if p.Alias == provider.Alias {
			cfg.Providers[i] = provider
			updated = true
			break
		}
	}
	if !updated {
		cfg.Providers = append(cfg.Providers, provider)
	}

	if err := config.SaveProviders(cfg); err != nil {
		return err
	}

	path, _ := config.SneakConfigDir()
	fmt.Printf("\nProvider '%s' saved → %s/providers.toml\n", provider.Alias, path)
	return nil
}

// --- Connection Testing ---

func testConnection(p config.Provider) error {
	switch p.Type {
	case "jira":
		return testJiraConnection(p)
	case "azure":
		return testAzureConnection(p)
	default:
		return fmt.Errorf("unsupported provider type: %s", p.Type)
	}
}

func testJiraConnection(p config.Provider) error {
	host := strings.TrimRight(p.Host, "/")
	url := host + "/rest/api/2/myself"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	req.SetBasicAuth(p.Username, p.Token)
	req.Header.Set("Accept", "application/json")

	return doTestRequest(req)
}

func testAzureConnection(p config.Provider) error {
	host := strings.TrimRight(p.Host, "/")
	url := host + "/_apis/profile/profiles/me?api-version=7.1-preview.1"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	// Azure DevOps PAT auth: Basic with empty user and PAT as password
	req.SetBasicAuth("", p.Token)
	req.Header.Set("Accept", "application/json")

	return doTestRequest(req)
}

func doTestRequest(req *http.Request) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// Consume body so connection can be reused
	io.Copy(io.Discard, resp.Body)
	return nil
}

// --- Prompt Helpers ---

func promptLine(reader *bufio.Reader, label string) (string, error) {
	fmt.Printf("%s: ", label)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func promptChoice(reader *bufio.Reader, label string, options []string) (string, error) {
	for {
		fmt.Printf("%s [%s]: ", label, strings.Join(options, "/"))
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("read input: %w", err)
		}
		choice := strings.TrimSpace(strings.ToLower(line))
		for _, opt := range options {
			if choice == opt {
				return opt, nil
			}
		}
		fmt.Printf("  invalid choice, please enter one of: %s\n", strings.Join(options, ", "))
	}
}

func readSecret() (string, error) {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		fmt.Println() // newline after hidden input
		if err != nil {
			return "", fmt.Errorf("read password: %w", err)
		}
		return string(b), nil
	}

	// Fallback for piped/non-terminal input
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}
