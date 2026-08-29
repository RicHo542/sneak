package cli

import (
	"fmt"

	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/git"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newStandupCmd() *cobra.Command {
	var (
		days      int
		multirepo bool
	)

	cmd := &cobra.Command{
		Use:   "sup",
		Short: "Get a standup summary of all your work",
		Long: `Get a standup summary of all your work.
This summary will be made for all repositories connected to sneak somehow.
The summary contains the git commit history of the given period, filtered to
the current user's git identity.

Use '--days' to control the look-back period (defaults to 1 day).
Use '--multirepo' to also scan for repos at the sneak projects' root dir.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStandupCmd(days, multirepo)
		},
	}

	cmd.Flags().IntVar(&days, "days", 1, "Number of days to look back for the summary.")
	cmd.Flags().BoolVarP(&multirepo, "multirepo", "m", false, "Whether to check for different repos at sneak root dirs.")

	return cmd
}

func runStandupCmd(
	days int, multirepo bool,
) error {

	repos, err := findAllKnownRepositories(multirepo)
	if err != nil {
		return err
	}

	since := fmt.Sprintf("%d days ago", days)
	gc := git.NewGitClient()

	summaries := make([]ui.RepoSummary, 0, len(repos))
	for _, repo := range repos {
		author, err := gc.GetGitUser(repo)
		if err != nil {
			ui.Printfln("Skipping '%s': %v", repo, err)
			continue
		}

		commits, err := gc.LogCommits(repo, since, author)
		if err != nil {
			return err
		}

		summaries = append(summaries, ui.RepoSummary{
			Path:    repo,
			Author:  author,
			Commits: commits,
		})
	}

	ui.PrintStandupSummary(summaries)
	return nil
}

func findAllKnownRepositories(multirepo bool) ([]string, error) {
	activeStates, err := config.DiscoverStates()
	if err != nil {
		return nil, err
	}

	var repos []string
	for _, s := range activeStates {
		stateRepos, err := DiscoverGitRepos(s.Dir, multirepo)
		if err != nil {
			return nil, err
		}

		repos = append(repos, stateRepos...)
	}

	return repos, nil
}
