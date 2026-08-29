package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type GitClient struct{}

// Commit is a single commit entry in a git history.
type Commit struct {
	Hash    string
	Author  string
	Date    string
	Message string
	// Type is the parsed conventional-commit type (feat, fix, ...) or "".
	Type string
	// Count is the number of merged commits represented by this entry; 1 for a
	// single commit.
	Count int
}

func NewGitClient() *GitClient {
	return &GitClient{}
}

// IsRepo checks if the given path is a valid git repository
func (c *GitClient) IsRepo(path string) bool {
	err := exec.Command("git", "-C", path, "rev-parse", "--git-dir").Run()
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

// GetGitUser returns the git user identifier (name, falling back to email) of the
// authenticated user within the given repository. It returns an empty string if no
// identity is configured.
func (c *GitClient) GetGitUser(repo string) (string, error) {
	for _, key := range []string{"user.name", "user.email"} {
		output, err := exec.Command("git", "-C", repo, "config", "--get", key).Output()
		if err != nil {
			continue
		}
		if name := strings.TrimSpace(string(output)); name != "" {
			return name, nil
		}
	}
	return "", fmt.Errorf("no git identity configured in '%s'", repo)
}

// LogCommits returns the structured commit history of the given repository within
// the time window, optionally filtered to a single author.
func (c *GitClient) LogCommits(repo, since, author string) ([]Commit, error) {
	args := []string{
		"-C", repo,
		"log",
		"--since=" + since,
		"--pretty=format:%H%x1f%an%x1f%ad%x1f%s",
		"--date=iso-strict",
	}
	if author != "" {
		args = append(args, "--author="+author)
	}

	output, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log in '%s': %v", repo, err)
	}

	var commits []Commit
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\x1f")
		if len(parts) != 4 {
			continue
		}

		commits = append(commits, Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Date:    parts[2],
			Message: parts[3],
			Type:    CommitType(parts[3]),
			Count:   1,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read git log output: %v", err)
	}

	return commits, nil
}
