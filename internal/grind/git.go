package grind

import (
	"fmt"
	"os/exec"
	"strings"
)

// CommitCount returns the number of commits on branch that are not in main.
// Uses git rev-list with --not main, matching GrindCLI's approach.
func CommitCount(bareRepo, branch string) (int, error) {
	cmd := exec.Command("git", "--git-dir="+bareRepo, "rev-list", "--count", branch, "--not", "main")
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("commit count: %w", err)
	}
	var n int
	_, err = fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &n)
	if err != nil {
		return 0, fmt.Errorf("parsing commit count: %w", err)
	}
	return n, nil
}

// FirstCommitDate returns the ISO-8601 datetime of the first commit on branch.
func FirstCommitDate(bareRepo, branch string) (string, error) {
	cmd := exec.Command("git", "--git-dir="+bareRepo, "log", "--reverse", "--format=%cI", branch)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("first commit date: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return "", nil
	}
	return strings.TrimSpace(lines[0]), nil
}

// LastCommitDate returns the ISO-8601 datetime of the most recent commit on branch.
func LastCommitDate(bareRepo, branch string) (string, error) {
	cmd := exec.Command("git", "--git-dir="+bareRepo, "log", "-1", "--format=%cI", branch)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("last commit date: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// HasUncommittedChanges checks whether the given worktree path has dirty or
// untracked files by running git status --porcelain.
func HasUncommittedChanges(worktreePath string) (bool, error) {
	cmd := exec.Command("git", "-C", worktreePath, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("checking uncommitted changes: %w", err)
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}
