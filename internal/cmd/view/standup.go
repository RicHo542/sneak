package view

import (
	"fmt"

	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/git"
	"github.com/richo542/sneak/internal/handlers"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newStandupCmd() *cobra.Command {
	var (
		days        int
		verbose     bool
		allBranches bool
	)

	cmd := &cobra.Command{
		Use:     "sup",
		Aliases: []string{"standup"},
		Short:   "Get a standup summary of all your work",
		Long: `Get a standup summary of all your work.
This summary will be made for all repositories connected to sneak somehow.
The summary contains the git commit history of the given period, filtered to
the current user's git identity.

Use '--days' to control the look-back period (defaults to 1 day).
Use '--all-branches' to include work from all local branches, not just the
currently checked-out one.
Use '--verbose' to also list repos without changes in the period.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStandupCmd(days, verbose, allBranches)
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 1, "Number of days to look back for the summary.")
	cmd.Flags().BoolVarP(&allBranches, "all-branches", "a", false, "Include work from all local branches.")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "List all found repos, including those without changes.")

	return cmd
}

type projectInfo struct {
	Dir   string
	Repos []string
}

func runStandupCmd(
	days int, verbose bool, allBranches bool,
) error {

	projects, err := findProjectRepositories()
	if err != nil {
		return err
	}

	since := fmt.Sprintf("%d days ago", days)
	gc := git.NewGitClient()

	summaries := make([]ui.ProjectSummary, 0, len(projects))
	var noChange []string
	for _, proj := range projects {
		project := ui.ProjectSummary{Root: proj.Dir}

		for _, repo := range proj.Repos {
			author, err := gc.GetGitUser(repo)
			if err != nil {
				if verbose {
					ui.Printfln("Skipping '%s': %v", repo, err)
				}
				continue
			}

			commits, err := gc.LogCommits(repo, since, author, allBranches)
			if err != nil {
				return err
			}

			// Dedup identical commits first (the same commit may be reached from
			// multiple branches with --all), then collapse repeated messages.
			commits = git.DedupByHash(commits)
			commits = git.CompactByMessage(commits)

			if len(commits) == 0 {
				if verbose {
					noChange = append(noChange, repo)
				}
				continue
			}

			project.Repos = append(project.Repos, ui.RepoSummary{
				Path:    repo,
				Author:  author,
				Commits: commits,
			})
		}

		if len(project.Repos) > 0 {
			summaries = append(summaries, project)
		}
	}

	ui.PrintStandupSummary(summaries, noChange)
	return nil
}

func findProjectRepositories() ([]projectInfo, error) {
	activeStates, err := config.DiscoverStates()
	if err != nil {
		return nil, err
	}

	projects := make([]projectInfo, 0, len(activeStates))
	for _, s := range activeStates {
		stateRepos, err := handlers.DiscoverGitRepos(s.Dir)
		if err != nil {
			return nil, err
		}

		if len(stateRepos) == 0 {
			continue
		}

		projects = append(projects, projectInfo{
			Dir:   s.Dir,
			Repos: stateRepos,
		})
	}

	return projects, nil
}
