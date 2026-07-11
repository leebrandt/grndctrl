package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestFindWorkspace verifies that FindWorkspace walks up directories correctly.
func TestFindWorkspace(t *testing.T) {
	t.Run("finds workspace from subdirectory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .grind.repo.git at the root
		bareRepoPath := filepath.Join(tmpDir, bareRepoName)
		if err := os.MkdirAll(bareRepoPath, 0755); err != nil {
			t.Fatalf("creating bare repo dir: %v", err)
		}

		// Create a nested subdirectory to search from
		subDir := filepath.Join(tmpDir, "a", "b", "c")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("creating subdir: %v", err)
		}

		got, err := FindWorkspace(subDir)
		if err != nil {
			t.Fatalf("FindWorkspace(%q) error: %v", subDir, err)
		}
		if got != tmpDir {
			t.Errorf("FindWorkspace(%q) = %q, want %q", subDir, got, tmpDir)
		}
	})

	t.Run("finds workspace from parent directory (worktree sibling)", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .grind.repo.git at tmpDir
		bareRepoPath := filepath.Join(tmpDir, bareRepoName)
		if err := os.MkdirAll(bareRepoPath, 0755); err != nil {
			t.Fatalf("creating bare repo dir: %v", err)
		}

		// Create a worktree directory as a sibling (child of tmpDir)
		worktreeDir := filepath.Join(tmpDir, "my-project")
		if err := os.MkdirAll(worktreeDir, 0755); err != nil {
			t.Fatalf("creating worktree dir: %v", err)
		}

		// Searching from the worktree should find tmpDir
		got, err := FindWorkspace(worktreeDir)
		if err != nil {
			t.Fatalf("FindWorkspace(%q) error: %v", worktreeDir, err)
		}
		if got != tmpDir {
			t.Errorf("FindWorkspace(%q) = %q, want %q", worktreeDir, got, tmpDir)
		}
	})

	t.Run("returns error when not in workspace", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := FindWorkspace(tmpDir)
		if err == nil {
			t.Error("expected error when not in a workspace")
		}
	})
}

// TestFindWorkspaceOrFlag tests the flag-or-auto-detect logic.
func TestFindWorkspaceOrFlag(t *testing.T) {
	t.Run("uses flag value when provided", func(t *testing.T) {
		tmpDir := t.TempDir()

		bareRepoPath := filepath.Join(tmpDir, bareRepoName)
		if err := os.MkdirAll(bareRepoPath, 0755); err != nil {
			t.Fatalf("creating bare repo dir: %v", err)
		}

		got, err := FindWorkspaceOrFlag(tmpDir)
		if err != nil {
			t.Fatalf("FindWorkspaceOrFlag(%q) error: %v", tmpDir, err)
		}
		if got != tmpDir {
			t.Errorf("FindWorkspaceOrFlag(%q) = %q, want %q", tmpDir, got, tmpDir)
		}
	})

	t.Run("errors when flag path has no workspace", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := FindWorkspaceOrFlag(tmpDir)
		if err == nil {
			t.Error("expected error for non-workspace path")
		}
	})

	t.Run("errors when flag path does not exist", func(t *testing.T) {
		_, err := FindWorkspaceOrFlag("/nonexistent/path")
		if err == nil {
			t.Error("expected error for nonexistent path")
		}
	})
}

// setupCollectProjectsFixture creates a temporary grind workspace with a real
// bare repo and linked worktrees containing .project.json files.
func setupCollectProjectsFixture(t *testing.T) (workspaceRoot string, cleanup func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "grind-collect-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}

	bareRepo := filepath.Join(tmpDir, bareRepoName)
	workspaceRoot = tmpDir

	// Initialize bare repo
	if err := exec.Command("git", "init", "--bare", bareRepo).Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git init --bare: %v", err)
	}

	// Create a working directory to set up branches
	workingDir := filepath.Join(tmpDir, ".working")
	if err := exec.Command("git", "init", workingDir).Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git init working: %v", err)
	}

	gitCfg := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", workingDir, "config"}, args...)...)
		cmd.Run()
	}
	gitCfg("user.name", "Test")
	gitCfg("user.email", "test@test.com")

	// Add bare repo as remote
	exec.Command("git", "-C", workingDir, "remote", "add", "origin", bareRepo).Run()

	// Create initial commit (creates default branch, typically master)
	f, _ := os.CreateTemp(workingDir, "init-*.txt")
	f.WriteString("initial")
	f.Close()
	exec.Command("git", "-C", workingDir, "add", ".").Run()
	exec.Command("git", "-C", workingDir, "commit", "-m", "initial").Run()
	// Rename default branch to main
	exec.Command("git", "-C", workingDir, "branch", "-m", "main").Run()
	exec.Command("git", "-C", workingDir, "push", "-u", "origin", "main").Run()

	// Create branches for two projects
	for _, branch := range []string{"project-alpha", "project-beta"} {
		exec.Command("git", "-C", workingDir, "checkout", "-b", branch).Run()
		f, _ := os.CreateTemp(workingDir, branch+"-*.txt")
		f.WriteString(branch)
		f.Close()
		exec.Command("git", "-C", workingDir, "add", ".").Run()
		exec.Command("git", "-C", workingDir, "commit", "-m", branch+" commit").Run()
		exec.Command("git", "-C", workingDir, "push", "-u", "origin", branch).Run()
		exec.Command("git", "-C", workingDir, "checkout", "main").Run()
	}

	// Create worktrees for each project branch
	worktrees := map[string]string{
		"project-alpha": filepath.Join(tmpDir, "project-alpha"),
		"project-beta":  filepath.Join(tmpDir, "project-beta"),
	}

	for branch, wtPath := range worktrees {
		cmd := exec.Command("git", "--git-dir="+bareRepo, "worktree", "add", wtPath, branch)
		if out, err := cmd.CombinedOutput(); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("git worktree add %s: %v\n%s", branch, err, out)
		}
	}

	// Create .project.json files in each worktree
	projectAlphaJSON := `{
		"name": "Project Alpha",
		"type": "app",
		"idea": "The alpha project",
		"billing": { "roundTo": "quarter-hour", "rate": 100 },
		"time": [
			{ "start": "2025-01-01T09:00:00Z", "end": "2025-01-01T11:00:00Z", "duration": 7200, "rounded": 7200 }
		]
	}`
	if err := os.WriteFile(filepath.Join(worktrees["project-alpha"], ".project.json"), []byte(projectAlphaJSON), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("writing .project.json for alpha: %v", err)
	}

	projectBetaJSON := `{
		"name": "Project Beta",
		"type": "cli",
		"idea": "The beta project",
		"billing": { "roundTo": "half-hour", "rate": 150 },
		"time": [
			{ "start": "2025-02-01T10:00:00Z", "end": "2025-02-01T12:00:00Z", "duration": 7200, "rounded": 7200, "invoiced": true }
		]
	}`
	if err := os.WriteFile(filepath.Join(worktrees["project-beta"], ".project.json"), []byte(projectBetaJSON), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("writing .project.json for beta: %v", err)
	}

	cleanup = func() {
		os.RemoveAll(tmpDir)
	}

	return workspaceRoot, cleanup
}

// setupGrindProjectsFixture creates a workspace with the real Grind layout:
// .project.json files live under each worktree's projects/<name>/ directory,
// matching Grind's write path.
func setupGrindProjectsFixture(t *testing.T) (workspaceRoot string, cleanup func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "grind-projects-dir-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}

	bareRepo := filepath.Join(tmpDir, bareRepoName)
	workspaceRoot = tmpDir

	if err := exec.Command("git", "init", "--bare", bareRepo).Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git init --bare: %v", err)
	}

	workingDir := filepath.Join(tmpDir, ".working")
	if err := exec.Command("git", "init", workingDir).Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git init working: %v", err)
	}

	gitCfg := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", workingDir, "config"}, args...)...)
		cmd.Run()
	}
	gitCfg("user.name", "Test")
	gitCfg("user.email", "test@test.com")

	exec.Command("git", "-C", workingDir, "remote", "add", "origin", bareRepo).Run()

	f, _ := os.CreateTemp(workingDir, "init-*.txt")
	f.WriteString("initial")
	f.Close()
	exec.Command("git", "-C", workingDir, "add", ".").Run()
	exec.Command("git", "-C", workingDir, "commit", "-m", "initial").Run()
	exec.Command("git", "-C", workingDir, "branch", "-m", "main").Run()
	exec.Command("git", "-C", workingDir, "push", "-u", "origin", "main").Run()

	exec.Command("git", "-C", workingDir, "checkout", "-b", "grind").Run()
	f2, _ := os.CreateTemp(workingDir, "grind-*.txt")
	f2.WriteString("grind")
	f2.Close()
	exec.Command("git", "-C", workingDir, "add", ".").Run()
	exec.Command("git", "-C", workingDir, "commit", "-m", "grind commit").Run()
	exec.Command("git", "-C", workingDir, "push", "-u", "origin", "grind").Run()

	grindDir := filepath.Join(tmpDir, "grind")
	cmd := exec.Command("git", "--git-dir="+bareRepo, "worktree", "add", grindDir, "grind")
	if out, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git worktree add grind: %v\n%s", err, out)
	}

	for _, branch := range []string{"project-alpha", "project-beta"} {
		exec.Command("git", "-C", workingDir, "checkout", "-b", branch).Run()
		bf, _ := os.CreateTemp(workingDir, branch+"-*.txt")
		bf.WriteString(branch)
		bf.Close()
		exec.Command("git", "-C", workingDir, "add", ".").Run()
		exec.Command("git", "-C", workingDir, "commit", "-m", branch+" commit").Run()
		exec.Command("git", "-C", workingDir, "push", "-u", "origin", branch).Run()
		exec.Command("git", "-C", workingDir, "checkout", "main").Run()

		wtPath := filepath.Join(tmpDir, branch)
		wtCmd := exec.Command("git", "--git-dir="+bareRepo, "worktree", "add", wtPath, branch)
		if out, err := wtCmd.CombinedOutput(); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("git worktree add %s: %v\n%s", branch, err, out)
		}
	}

	alphaDir := filepath.Join(tmpDir, "project-alpha", "projects", "project-alpha")
	if err := os.MkdirAll(alphaDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("creating alpha dir: %v", err)
	}
	alphaJSON := `{
		"name": "Project Alpha",
		"type": "app",
		"idea": "The alpha project",
		"billing": { "roundTo": "quarter-hour", "rate": 100 },
		"time": [
			{ "start": "2025-01-01T09:00:00Z", "end": "2025-01-01T11:00:00Z", "duration": 7200, "rounded": 7200 }
		]
	}`
	if err := os.WriteFile(filepath.Join(alphaDir, ".project.json"), []byte(alphaJSON), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("writing alpha .project.json: %v", err)
	}

	betaDir := filepath.Join(tmpDir, "project-beta", "projects", "project-beta")
	if err := os.MkdirAll(betaDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("creating beta dir: %v", err)
	}
	betaJSON := `{
		"name": "Project Beta",
		"type": "cli",
		"idea": "The beta project",
		"billing": { "roundTo": "half-hour", "rate": 150 },
		"time": [
			{ "start": "2025-02-01T10:00:00Z", "end": "2025-02-01T12:00:00Z", "duration": 7200, "rounded": 7200, "invoiced": true }
		]
	}`
	if err := os.WriteFile(filepath.Join(betaDir, ".project.json"), []byte(betaJSON), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("writing beta .project.json: %v", err)
	}

	gammaDir := filepath.Join(grindDir, "projects", "project-gamma")
	if err := os.MkdirAll(gammaDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("creating gamma dir: %v", err)
	}
	gammaJSON := `{
		"name": "Project Gamma",
		"type": "lib",
		"idea": "The gamma project (cancelled — no worktree)",
		"billing": { "roundTo": "hour", "rate": 200 },
		"time": []
	}`
	if err := os.WriteFile(filepath.Join(gammaDir, ".project.json"), []byte(gammaJSON), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("writing gamma .project.json: %v", err)
	}

	cleanup = func() {
		os.RemoveAll(tmpDir)
	}

	return workspaceRoot, cleanup
}

func TestCollectProjectInfosFromGrindProjectsDir(t *testing.T) {
	workspaceRoot, cleanup := setupGrindProjectsFixture(t)
	defer cleanup()

	infos, err := CollectProjectInfos(workspaceRoot)
	if err != nil {
		t.Fatalf("CollectProjectInfos() error: %v", err)
	}

	if len(infos) != 2 {
		t.Fatalf("CollectProjectInfos() returned %d projects, want 2 (gamma excluded — no worktree)", len(infos))
	}

	names := make(map[string]bool)
	for _, info := range infos {
		names[info.Config.Name] = true
	}
	if !names["Project Alpha"] {
		t.Error("CollectProjectInfos() missing 'Project Alpha'")
	}
	if !names["Project Beta"] {
		t.Error("CollectProjectInfos() missing 'Project Beta'")
	}
	if names["Project Gamma"] {
		t.Error("CollectProjectInfos() should exclude 'Project Gamma' (no worktree)")
	}

	for _, info := range infos {
		if info.Config.Name == "Project Alpha" {
			if info.Config.Billing.Rate != 100 {
				t.Errorf("Project Alpha rate = %v, want 100", info.Config.Billing.Rate)
			}
		}
		if info.Config.Name == "Project Beta" {
			if info.Config.Billing.Rate != 150 {
				t.Errorf("Project Beta rate = %v, want 150", info.Config.Billing.Rate)
			}
		}
	}
}

func TestCollectProjects(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	workspaceRoot, cleanup := setupCollectProjectsFixture(t)
	defer cleanup()

	projects, err := CollectProjects(workspaceRoot)
	if err != nil {
		t.Fatalf("CollectProjects() error: %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("CollectProjects() returned %d projects, want 2", len(projects))
	}

	// Verify both projects are found by name
	names := make(map[string]bool)
	for _, p := range projects {
		names[p.Name] = true
	}
	if !names["Project Alpha"] {
		t.Error("CollectProjects() missing 'Project Alpha'")
	}
	if !names["Project Beta"] {
		t.Error("CollectProjects() missing 'Project Beta'")
	}

	// Verify project data was read correctly
	for _, p := range projects {
		if p.Name == "Project Alpha" {
			if p.Billing.Rate != 100 {
				t.Errorf("Project Alpha rate = %v, want 100", p.Billing.Rate)
			}
			if len(p.Time) != 1 {
				t.Errorf("Project Alpha has %d sessions, want 1", len(p.Time))
			}
			if p.TotalSeconds() != 7200 {
				t.Errorf("Project Alpha TotalSeconds = %d, want 7200", p.TotalSeconds())
			}
		}
		if p.Name == "Project Beta" {
			if p.Billing.Rate != 150 {
				t.Errorf("Project Beta rate = %v, want 150", p.Billing.Rate)
			}
			if p.BilledSeconds() != 7200 {
				t.Errorf("Project Beta BilledSeconds = %d, want 7200", p.BilledSeconds())
			}
		}
	}
}
