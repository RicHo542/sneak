package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/richo542/sneak/internal/git"
)

func DiscoverGitRepos(dir string) ([]string, error) {
	gc := git.NewGitClient()

	var repos []string
	if gc.IsRepo(dir) {
		repos = append(repos, dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s: %w", dir, err)
	}

	for _, candidate := range entries {
		if !candidate.IsDir() {
			continue
		}

		candidatePath := filepath.Join(dir, candidate.Name())
		if gc.IsRepo(candidatePath) {
			repos = append(repos, candidatePath)
		}
	}

	return repos, nil
}
