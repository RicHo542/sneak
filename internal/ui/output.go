package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/richo542/sneak/internal/config"
)

const (
	ColorTeal  = "\033[38;2;70;255;200m"
	ColorDark  = "\033[38;2;26;38;50m"
	ColorWhite = "\033[1;37m"
	ColorGray  = "\033[38;2;126;140;155m"
	ColorReset = "\033[0m"
)

func PrintBanner() {
	textArt := ColorWhite + `
  ███████╗███╗   ██╗███████╗███████╗██╗  ██╗
  ██╔════╝████╗  ██║██╔════╝██╔══██╗██║ ██║
  ███████╗██╔██╗ ██║█████╗  ███████║██║██║
  ╚════██║██║╚██╗██║██╔══╝  ██╔══██║██║ ██║
  ███████║██║ ╚████║███████╗██║  ██║██║  ██║
  ╚══════╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝` + ColorTeal + `__` + ColorReset + `
`

	tagline := ColorGray + "   	 sneak: minimize red tape_" + ColorReset + "\n"

	fmt.Print(textArt)
	fmt.Println(tagline)
}

func Printfln(format string, a ...any) {
	fmt.Printf(format+"\n", a...)
}

func PrintTableOfItems(items []config.CacheItem) {
	fmt.Printf("%-12s  %-10s  %-12s  %-16s  %s\n", "KEY", "ASSIGNED", "TYPE", "STATUS", "SUMMARY")
	fmt.Println(strings.Repeat("-", 100))

	for _, item := range items {
		assignFlag := ""
		if item.Assignee != "" {
			assignFlag = "x"
		}
		summary := item.Summary
		if len(summary) > 40 {
			summary = item.Summary[:40]
		}
		fmt.Printf("%-12s  %-10s  %-12s  %-16s  %s\n", item.Key, assignFlag, item.Type, item.Status, summary)
	}
}

func PrintActiveTaskTable(items []config.ActiveTask) {
	fmt.Printf("%-12s  %-16s  %-16s  %-16s  %s\n", "KEY", "STATUS", "ACTIVATED", "BRANCH", "SUMMARY")
	fmt.Println(strings.Repeat("-", 100))

	for _, item := range items {
		summary := item.Summary
		if len(summary) > 40 {
			summary = item.Summary[:40]
		}

		fmt.Printf("%-12s  %-16s  %-16s  %-16s  %s\n", item.Key, item.Status, timeAgo(item.ActivatedAt), item.Branch, summary)
	}
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	if d < 0 {
		d = 0
	}

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
