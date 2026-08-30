package ui

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenBrowser launches the default browser for the given URL on the current
// platform. It returns an error if no suitable opener can be resolved or the
// command fails to start.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	// Detach from the process so a short-lived CLI never blocks on the browser.
	go cmd.Wait()
	return nil
}

// OpenBrowserOrPrint opens the URL in a browser when possible, otherwise prints
// the URL so the user can open it manually. colors are applied via ui.Color.
func OpenBrowserOrPrint(url string) {
	if err := OpenBrowser(url); err != nil {
		fmt.Printf("%sCould not open browser: %v%s\n", Color(ColorGray), err, Color(ColorReset))
	}
	fmt.Println(url)
}
