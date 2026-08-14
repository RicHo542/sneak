package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type GitClient struct{}

func NewGitClient() *GitClient {
	return &GitClient{}
}

// IsRepo checks if the current directory is a
// valid git repository
func (c *GitClient) IsRepo() bool {
	err := exec.Command("git", "rev-parse", "--git-dir").Run()
	return err == nil
}

// CurrentBranchName retrieves the current branch name active in the
// current working directory
func (c *GitClient) CurrentBranchName() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// CreateBranch creates a new branch in the current git context.
// It verifies if the branch name is already taken and returns an
// error in case the name is not available.
func (c *GitClient) CreateBranch(branchName string) error {
	if c.BranchExists(branchName) {
		return fmt.Errorf("branch '%s' already exists.", branchName)
	}

	command := exec.Command("git", "switch", "-c", branchName)
	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	return command.Run()
}

// BranchExists checks if a given branch name already exists in the
// referenced repository
func (c *GitClient) BranchExists(branchName string) bool {
	err := exec.Command("git", "rev-parse", "--verify", branchName).Run()
	return err == nil
}

// IsWorkingTreeClean checks if the current git working tree is clean from changes
// by using 'git status --porcelain'
func (c *GitClient) IsWorkingTreeClean() (bool, error) {
	output, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return false, fmt.Errorf("failed to get git status: %v", err)
	}

	return len(output) == 0, nil
}
