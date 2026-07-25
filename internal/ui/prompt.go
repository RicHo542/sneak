package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func PromptLine(reader *bufio.Reader, label string) (string, error) {
	fmt.Printf("%s: ", label)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
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
