package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// EditText opens the user's editor pre-filled with banner and initial text.
// Banner lines (lines starting with '#') are stripped from the result, so the
// caller can safely display instructions that the user sees while editing but
// that are never persisted.
func EditText(initial string, banner string) (string, error) {
	editor := "vi"
	if e, ok := os.LookupEnv("EDITOR"); ok && e != "" {
		editor = e
	}

	var content string
	if banner != "" {
		content = banner + "\n\n" + initial
	} else {
		content = initial
	}

	tmpFile, err := os.CreateTemp("", "sneak-note-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	raw, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read editor output: %w", err)
	}

	// Strip leading banner lines (# comments) and trim.
	lines := strings.Split(string(raw), "\n")
	result := make([]string, 0, len(lines))
	inBanner := true
	for _, line := range lines {
		if inBanner && strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		inBanner = false
		result = append(result, line)
	}

	return strings.TrimSpace(strings.Join(result, "\n")), nil
}
