package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/richo542/sneak/internal/client/objects"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/git"
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

// PrintWorkItemDetail renders the full detail view used by 'sneak describe'.
func PrintWorkItemDetail(detail *objects.WorkItemDetail) {
	fmt.Printf("%s: %s\n", detail.Key, detail.Name)
	if detail.URL != "" {
		fmt.Println(detail.URL)
	}
	fmt.Println()

	fmt.Println("Description:")
	if strings.TrimSpace(detail.Description) == "" {
		fmt.Println("  (none)")
	} else {
		for _, line := range strings.Split(detail.Description, "\n") {
			fmt.Printf("  %s\n", line)
		}
	}
	fmt.Println()

	createdBy := detail.CreatedBy
	if createdBy == "" {
		createdBy = "(unknown)"
	}
	owner := detail.Owner
	if owner == "" {
		owner = "(unassigned)"
	}
	iteration := detail.IterationPath
	if iteration == "" {
		iteration = "(unknown)"
	}

	fmt.Printf("Created:   %s by %s\n", detail.CreatedAt, createdBy)
	fmt.Printf("Iteration: %s\n", iteration)
	fmt.Printf("Owner:     %s\n", owner)
	fmt.Println()

	if detail.TotalComments > len(detail.Comments) {
		fmt.Printf("Comments (showing last %d of %d):\n", len(detail.Comments), detail.TotalComments)
	} else {
		fmt.Printf("Comments (%d):\n", len(detail.Comments))
	}

	if len(detail.Comments) == 0 {
		fmt.Println("  (none)")
		return
	}

	for _, c := range detail.Comments {
		author := c.Author
		if author == "" {
			author = "(unknown)"
		}
		fmt.Printf("  [%s] %s: %s\n", c.CreatedAt, author, c.Body)
	}
}

// RepoSummary is a standup summary entry for a single repository.
type RepoSummary struct {
	Path    string
	Author  string
	Commits []git.Commit
}

// PrintStandupSummary renders a per-repository git commit overview.
func PrintStandupSummary(summaries []RepoSummary) {
	for _, s := range summaries {
		fmt.Printf("\n%s%s%s\n", ColorTeal, s.Path, ColorReset)
		fmt.Println(strings.Repeat("-", 100))

		if len(s.Commits) == 0 {
			Printfln("%sNo commits by %s in this period.%s", ColorGray, s.Author, ColorReset)
			continue
		}

		for _, c := range s.Commits {
			shortHash := c.Hash
			if len(shortHash) > 7 {
				shortHash = shortHash[:7]
			}
			date := c.Date
			if len(date) > 19 {
				date = date[:19]
			}
			fmt.Printf("%s%s %s %s%s\n", ColorGray, date, shortHash, c.Message, ColorReset)
		}
	}
}

func timeAgo(t time.Time) string {
	d := time.Since(t)

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
