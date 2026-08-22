package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/richo542/sneak/internal/config"
	"golang.org/x/term"

	"charm.land/huh/v2"
)

type SelectItem struct {
	Key   string
	Label string
}

var HuhAppTheme = huh.ThemeFunc(huh.ThemeBase16)

func NewThemedAppForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).WithTheme(HuhAppTheme)
}

func PromptLine(reader *bufio.Reader, label string) (string, error) {
	fmt.Printf("%s: ", label)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func PromptSelect(reader *bufio.Reader, label string, items []SelectItem) (string, error) {
	for {
		fmt.Printf("%s:\n", label)
		for i, item := range items {
			fmt.Printf("  %d) %s\n", i+1, item.Label)
		}
		fmt.Print("Enter number or key: ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("read input: %w", err)
		}
		choice := strings.TrimSpace(line)

		// Try numeric selection
		for i, item := range items {
			if choice == fmt.Sprintf("%d", i+1) {
				return item.Key, nil
			}
		}

		// Try matching by key
		for _, item := range items {
			if strings.EqualFold(choice, item.Key) {
				return item.Key, nil
			}
		}

		fmt.Printf("  invalid choice, please enter a number (1-%d) or key\n", len(items))
	}
}

func PromptChoice(reader *bufio.Reader, label string, options []string) (string, error) {
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

func ReadSecret() (string, error) {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		fmt.Println()
		if err != nil {
			return "", fmt.Errorf("read password: %w", err)
		}
		return string(b), nil
	}

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func InteractiveSelectItem(items []*config.CacheItem) ([]*config.CacheItem, error) {
	var selection []*config.CacheItem

	var options []huh.Option[*config.CacheItem]
	for _, item := range items {
		options = append(options, huh.NewOption(
			fmt.Sprintf("%s: [%s] %s", item.Key, item.Type, item.Summary),
			item,
		))
	}
	// +1 for header
	selectionHeight := min(12, len(options)+1)

	selectBox := NewThemedAppForm(
		huh.NewGroup(
			huh.NewMultiSelect[*config.CacheItem]().
				Options(options...).
				Title("Available Work Items:").
				Height(selectionHeight).
				Value(&selection),
		),
	).WithShowHelp(true)

	if err := selectBox.Run(); err != nil {
		return nil, err
	}

	return selection, nil
}
