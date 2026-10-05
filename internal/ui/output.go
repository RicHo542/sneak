package ui

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/richo542/sneak/internal/client/objects"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/git"
	"github.com/richo542/sneak/internal/todos"
	"golang.org/x/term"
)

const (
	ColorTeal  = "\033[38;2;70;255;200m"
	ColorDark  = "\033[38;2;26;38;50m"
	ColorWhite = "\033[1;37m"
	ColorGray  = "\033[38;2;126;140;155m"
	ColorReset = "\033[0m"
	ColorRed   = "\033[38;2;255;100;100m"
	Dot        = "●"
)

// useColor reports whether ANSI colors should be emitted: only when output is an
// interactive terminal and the NO_COLOR convention has not opted out.
func useColor() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Color returns the ANSI escape matching code when colors are enabled, or an
// empty string otherwise.
func Color(code string) string {
	if useColor() {
		return code
	}
	return ""
}

// colorForCommitType returns the display color for a commit type. As a light
// touch, only 'feat' and 'fix' are highlighted; everything else stays gray.
func colorForCommitType(commitType string) string {
	switch commitType {
	case "feat":
		return ColorTeal
	case "fix":
		return ColorRed
	default:
		return ColorGray
	}
}

func PrintBanner() {
	textArt := Color(ColorWhite) + `
  ███████╗███╗   ██╗███████╗███████╗██╗  ██╗
  ██╔════╝████╗  ██║██╔════╝██╔══██╗██║ ██║
  ███████╗██╔██╗ ██║█████╗  ███████║██║██║
  ╚════██║██║╚██╗██║██╔══╝  ██╔══██║██║ ██║
  ███████║██║ ╚████║███████╗██║  ██║██║  ██║
  ╚══════╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝` + Color(ColorTeal) + `__` + Color(ColorReset) + `
`

	tagline := Color(ColorGray) + "   	 sneak: minimize red tape_" + Color(ColorReset) + "\n"

	fmt.Print(textArt)
	fmt.Println(tagline)
}

func Printfln(format string, a ...any) {
	fmt.Printf(format+"\n", a...)
}

func ColoredString(v string, color string) string {
	return Color(color) + v + Color(ColorReset)
}

func PrintTableOfProviderItems(items []config.CacheItem) {
	fmt.Printf("%-12s  %-10s  %-12s  %-16s  %s\n", "KEY", "ASSIGNED", "TYPE", "STATUS", "SUMMARY")
	fmt.Println(strings.Repeat("-", 100))

	for _, item := range items {
		assignFlag := ""
		if item.Assignee != "" {
			assignFlag = Dot
		}
		summary := item.Summary
		if len(summary) > 40 {
			summary = item.Summary[:40]
		}
		fmt.Printf("%-12s  %-10s  %-12s  %-16s  %s\n", item.Key, assignFlag, item.Type, item.Status, summary)
	}
}

func PrintTableOfTodos(todos []*todos.Todo) {
	fmt.Printf("%-8s  %-4s  %-6s  %-10s  %s\n", "KEY", "PIN", "STATUS", "AGE", "TITLE")
	fmt.Println(strings.Repeat("-", 100))

	for _, item := range todos {
		age := time.Since(item.CreatedAt).Truncate(time.Minute)
		pinMarker := ""
		if item.Pin {
			pinMarker = Dot
		}
		title := item.Title
		if len(title) > 50 {
			title = fmt.Sprintf("%s...", item.Title[:50])
		}
		fmt.Printf("%-8s  %-4s  %-6s  %-10s  %s\n", item.Key, pinMarker, item.Status, age, title)
	}
	fmt.Println()
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

// PrintTodoDetail renders the full detail view of a local todo, used by
// 'sneak describe' for sn: refs.
func PrintTodoDetail(item *todos.Todo) {
	Printfln(ColoredString("Key:    ", ColorTeal)+"%s", item.Key)
	Printfln(ColoredString("Title:  ", ColorTeal)+"%s", item.Title)
	fmt.Println()

	Printfln(ColoredString("Status:   ", ColorTeal)+"%s", item.Status)
	Printfln(ColoredString("Created:  ", ColorTeal)+"%s", item.CreatedAt.Format("2006-01-02 15:04"))
	if item.StartedAt != nil {
		Printfln(ColoredString("Started:  ", ColorTeal)+"%s", item.StartedAt.Format("2006-01-02 15:04"))
	}
	if item.ClosedAt != nil {
		Printfln(ColoredString("Closed:   ", ColorTeal)+"%s", item.ClosedAt.Format("2006-01-02 15:04"))
	}
	if item.Pin {
		fmt.Println(ColoredString("Pinned:   ", ColorTeal) + "yes")
	}
	if len(item.Labels) > 0 {
		Printfln(ColoredString("Labels:   ", ColorTeal)+"%s", strings.Join(item.Labels, ", "))
	}
	fmt.Println()

	fmt.Println(ColoredString("Description:", ColorTeal))
	desc := strings.TrimSpace(item.Description)
	if desc == "" {
		fmt.Println("  (none)")
	} else {
		lines := strings.Split(desc, "\n")
		limit := 20
		for i := 0; i < len(lines) && i < limit; i++ {
			Printfln("  %s", lines[i])
		}
		if len(lines) > limit {
			Printfln("  ... (%d more lines)", len(lines)-limit)
			Printfln("  use 'sneak note %s' to see the full description/notes", item.Key)
		}
	}
	fmt.Println()

	Printfln(ColoredString("Comments (%d):", ColorTeal), len(item.Comments))
	if len(item.Comments) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, n := range item.Comments {
			at := ""
			if n.At != nil {
				at = n.At.Format("2006-01-02 15:04")
			}
			Printfln("  [%s] %s", at, n.Text)
		}
	}
}

// RepoSummary is a standup summary entry for a single repository.
type RepoSummary struct {
	Path    string
	Author  string
	Commits []git.Commit
}

// ProjectSummary groups the repository summaries of a single sneak project.
type ProjectSummary struct {
	Root  string
	Repos []RepoSummary
}

// PrintStandupSummary renders a per-project, per-repository git commit overview.
func PrintStandupSummary(projects []ProjectSummary, noChange []string) {
	for _, p := range projects {
		// Multi-repo projects show a project header, single-repo projects are
		// rendered directly to keep the common case compact.
		if len(p.Repos) > 1 {
			fmt.Printf("\n%s%s%s\n", Color(ColorWhite), p.Root, Color(ColorReset))
		}

		for _, s := range p.Repos {
			if len(p.Repos) > 1 && s.Path != p.Root {
				fmt.Printf("\n  %s%s%s\n", Color(ColorTeal), s.Path, Color(ColorReset))
			} else {
				fmt.Printf("\n%s%s%s\n", Color(ColorTeal), s.Path, Color(ColorReset))
			}
			fmt.Println(strings.Repeat("-", 100))

			for _, c := range s.Commits {
				shortHash := c.Hash
				if len(shortHash) > 7 {
					shortHash = shortHash[:7]
				}
				date := c.Date
				if len(date) > 19 {
					date = date[:19]
				}

				commitType := c.Type
				if commitType == "" {
					commitType = "-"
				}

				message := c.Message
				if c.Type != "" {
					message = git.StripTypePrefix(message)
				}
				if c.Count > 1 {
					message = fmt.Sprintf("%s (x%d)", message, c.Count)
				}

				typeColor := Color(colorForCommitType(commitType))
				fmt.Printf("%s%s  %s%-8s%s %s  %s%s\n",
					Color(ColorGray), date,
					typeColor, commitType, Color(ColorReset),
					shortHash, message, Color(ColorReset),
				)
			}
		}
	}

	if len(noChange) > 0 {
		fmt.Printf("\n%sNo changes in:%s\n", Color(ColorWhite), Color(ColorReset))
		for _, repo := range noChange {
			fmt.Printf("  %s%s%s\n", Color(ColorGray), repo, Color(ColorReset))
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

// TodoGroup holds local todos for a single project scope.
type TodoGroup struct {
	Project string
	Path    string
	Items   []*todos.Todo
}

// ProviderGroup holds cached provider items for a single project.
type ProviderGroup struct {
	Project string
	Path    string
	Items   []config.CacheItem
	Age     time.Duration
}

// shortPathMax is the default width of the PATH column in 'sneak all' tables.
const shortPathMax = 30

// PrintAllSummary renders 'sneak all' as two flat tables (provider items and
// local todos) with a trailing PATH column. Provider cache ages are listed per
// project under the provider table.
func PrintAllSummary(todoGroups []TodoGroup, providerGroups []ProviderGroup) {
	if len(providerGroups) == 0 && len(todoGroups) == 0 {
		fmt.Println("No work items or todos found.")
		return
	}

	if len(providerGroups) > 0 {
		type providerRow struct {
			item config.CacheItem
			path string
		}
		var rows []providerRow
		for _, pg := range providerGroups {
			for _, item := range pg.Items {
				rows = append(rows, providerRow{item: item, path: pg.Path})
			}
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].path != rows[j].path {
				return rows[i].path < rows[j].path
			}
			return rows[i].item.Key < rows[j].item.Key
		})

		fmt.Printf("%-12s  %-10s  %-12s  %-16s  %-32s  %s\n",
			"KEY", "ASSIGNED", "TYPE", "STATUS", "PATH", "SUMMARY")
		fmt.Println(strings.Repeat("-", 100))

		for _, r := range rows {
			assignFlag := ""
			if r.item.Assignee != "" {
				assignFlag = Dot
			}
			summary := r.item.Summary
			if len(summary) > 40 {
				summary = r.item.Summary[:40]
			}
			fmt.Printf("%-12s  %-10s  %-12s  %-16s  %-32s  %s\n",
				r.item.Key, assignFlag, r.item.Type, r.item.Status,
				shortPath(r.path, shortPathMax), summary)
		}

		fmt.Println()
		fmt.Printf("%d provider item(s) across %d project(s)\n", len(rows), len(providerGroups))
		sort.Slice(providerGroups, func(i, j int) bool {
			return providerGroups[i].Path < providerGroups[j].Path
		})
		for _, pg := range providerGroups {
			fmt.Printf("  %-40s  fetched %s ago\n", pg.Path, pg.Age)
		}
	}

	if len(todoGroups) > 0 {
		type todoRow struct {
			item *todos.Todo
			path string
		}
		var rows []todoRow
		for _, tg := range todoGroups {
			for _, item := range tg.Items {
				rows = append(rows, todoRow{item: item, path: tg.Path})
			}
		}
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].path != rows[j].path {
				return rows[i].path < rows[j].path
			}
			if rows[i].item.Pin != rows[j].item.Pin {
				return rows[i].item.Pin
			}
			return rows[i].item.CreatedAt.Before(rows[j].item.CreatedAt)
		})

		if len(providerGroups) > 0 {
			fmt.Println()
		}

		fmt.Printf("%-8s  %-4s  %-6s  %-10s  %-46s  %s\n",
			"KEY", "PIN", "STATUS", "AGE", "TITLE", "PATH")
		fmt.Println(strings.Repeat("-", 100))

		for _, r := range rows {
			age := time.Since(r.item.CreatedAt).Truncate(time.Minute)
			pinMarker := ""
			if r.item.Pin {
				pinMarker = Dot
			}
			title := r.item.Title
			if len(title) > 46 {
				title = r.item.Title[:46]
			}
			fmt.Printf("%-8s  %-4s  %-6s  %-10s  %-46s  %s\n",
				r.item.Key, pinMarker, r.item.Status, age, title,
				shortPath(r.path, shortPathMax))
		}

		fmt.Println()
		fmt.Printf("%d local todo(s) across %d project(s)\n", len(rows), len(todoGroups))
	}
}

// shortPath renders a project path for a PATH column: the full path when it
// fits within max characters, otherwise the trailing max characters prefixed
// with "...".
func shortPath(path string, max int) string {
	if len(path) <= max {
		return path
	}
	return "..." + path[len(path)-max:]
}
