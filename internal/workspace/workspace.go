// Package workspace handles discovery of Grind workspaces by locating the
// .grind.repo.git bare repository directory.
package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/leebrandt/grndctrl/internal/grind"
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

// CollectProjects discovers all project worktrees by running
// git worktree list against the bare repo, then loads .project.json
// from each worktree. Worktrees without a valid .project.json are silently
// skipped.
func CollectProjects(workspaceRoot string) ([]grind.ProjectConfig, error) {
	bareRepo := filepath.Join(workspaceRoot, bareRepoName)

	cmd := exec.Command("git", "--git-dir="+bareRepo, "worktree", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing worktrees: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var projects []grind.ProjectConfig

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		worktreePath := fields[0]

		projectFilePath := filepath.Join(worktreePath, ".project.json")
		if _, err := os.Stat(projectFilePath); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(projectFilePath)
		if err != nil {
			continue // skip unreadable files
		}

		var project grind.ProjectConfig
		if err := json.Unmarshal(data, &project); err != nil {
			continue // skip malformed files
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// ProjectInfo enriches a ProjectConfig with its worktree path and branch name.
type ProjectInfo struct {
	Config       grind.ProjectConfig
	WorktreePath string
	Branch       string
}

// CollectProjectInfos returns enriched project information including
// worktree path and branch name parsed from git worktree list output.
// Worktrees without a valid .project.json are silently skipped.
func CollectProjectInfos(workspaceRoot string) ([]ProjectInfo, error) {
	bareRepo := filepath.Join(workspaceRoot, bareRepoName)

	cmd := exec.Command("git", "--git-dir="+bareRepo, "worktree", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing worktrees: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var infos []ProjectInfo

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		worktreePath := fields[0]

		branch := strings.Trim(fields[2], "[]()")
		branch = strings.TrimPrefix(branch, "refs/heads/")

		projectFilePath := filepath.Join(worktreePath, ".project.json")
		if _, err := os.Stat(projectFilePath); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(projectFilePath)
		if err != nil {
			continue
		}

		var project grind.ProjectConfig
		if err := json.Unmarshal(data, &project); err != nil {
			continue
		}

		infos = append(infos, ProjectInfo{
			Config:       project,
			WorktreePath: worktreePath,
			Branch:       branch,
		})
	}

	return infos, nil
}
