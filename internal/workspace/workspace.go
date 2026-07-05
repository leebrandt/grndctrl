// Package workspace handles discovery of Grind workspaces by locating the
// .grind.repo.git bare repository directory.
package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const bareRepoName = ".grind.repo.git"

// FindWorkspace walks up from startDir looking for .grind.repo.git.
// Returns the absolute path to the workspace root (the parent directory
// containing .grind.repo.git), or an error if not found.
func FindWorkspace(startDir string) (string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}

	current := abs
	for {
		bareRepoPath := filepath.Join(current, bareRepoName)
		if info, err := os.Stat(bareRepoPath); err == nil && info.IsDir() {
			return current, nil
		}

		// Also check one level up (worktrees are siblings to the bare repo)
		parent := filepath.Dir(current)
		parentBareRepo := filepath.Join(parent, bareRepoName)
		if info, err := os.Stat(parentBareRepo); err == nil && info.IsDir() {
			return parent, nil
		}

		// Reached root without finding anything
		if parent == current {
			return "", errors.New("not in a grind workspace: could not find .grind.repo.git")
		}

		current = parent
	}
}

// FindWorkspaceOrFlag checks the --workspace flag value first, then falls back
// to auto-detection from the current working directory.
func FindWorkspaceOrFlag(flagValue string) (string, error) {
	if flagValue != "" {
		abs, err := filepath.Abs(flagValue)
		if err != nil {
			return "", fmt.Errorf("resolving --workspace path: %w", err)
		}
		// Verify the flag path actually contains a grind workspace.
		bareRepo := filepath.Join(abs, bareRepoName)
		if info, err := os.Stat(bareRepo); err != nil || !info.IsDir() {
			return "", fmt.Errorf("--workspace %s does not contain a grind workspace (no .grind.repo.git found)", flagValue)
		}
		return abs, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}
	return FindWorkspace(cwd)
}
