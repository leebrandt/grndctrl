package grind

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupBareRepo creates a bare git repository at barePath, adds a main branch
// with one commit, and a feature branch with additional commits. Returns the
// bare repo path and a cleanup function.
func setupBareRepo(t *testing.T) (bareRepo string, cleanup func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "grind-git-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}

	bareRepo = filepath.Join(tmpDir, ".grind.repo.git")
	workingDir := filepath.Join(tmpDir, "working")

	// Init bare repo
	if err := exec.Command("git", "init", "--bare", bareRepo).Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git init --bare: %v", err)
	}

	// Init working directory
	if err := exec.Command("git", "init", workingDir).Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git init working: %v", err)
	}

	// Configure user for commits
	gitConfig := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", workingDir, "config"}, args...)...)
		if err := cmd.Run(); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("git config %v: %v", args, err)
		}
	}
	gitConfig("user.name", "Test")
	gitConfig("user.email", "test@test.com")

	// Add bare repo as remote
	if err := exec.Command("git", "-C", workingDir, "remote", "add", "origin", bareRepo).Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git remote add: %v", err)
	}

	// Helper to add a commit on current branch
	commit := func(msg string) {
		f, err := os.CreateTemp(workingDir, "file-*.txt")
		if err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("creating temp file: %v", err)
		}
		f.WriteString(msg)
		f.Close()

		exec.Command("git", "-C", workingDir, "add", ".").Run()
		if err := exec.Command("git", "-C", workingDir, "commit", "-m", msg).Run(); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("git commit: %v", err)
		}
	}

	// Make initial commit (creates default branch, typically master)
	commit("initial commit")

	// Rename default branch to main
	exec.Command("git", "-C", workingDir, "branch", "-m", "main").Run()

	// Set upstream and push main to bare repo
	if err := exec.Command("git", "-C", workingDir, "push", "-u", "origin", "main").Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git push main: %v", err)
	}

	// Create and switch to feature branch
	if err := exec.Command("git", "-C", workingDir, "checkout", "-b", "feature-one").Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git checkout -b feature-one: %v", err)
	}

	// Add two commits on feature
	commit("feature commit 1")
	commit("feature commit 2")

	// Push feature to bare repo
	if err := exec.Command("git", "-C", workingDir, "push", "-u", "origin", "feature-one").Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git push feature-one: %v", err)
	}

	cleanup = func() {
		os.RemoveAll(tmpDir)
	}

	return bareRepo, cleanup
}

// setupWorktree creates a non-bare git repo at worktreePath for testing
// HasUncommittedChanges.
func setupWorktree(t *testing.T) (worktreePath string, cleanup func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "grind-worktree-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}

	worktreePath = filepath.Join(tmpDir, "worktree")

	// Init a non-bare repo
	cmd := exec.Command("git", "init", worktreePath)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("git init: %v", err)
	}

	// Configure user
	exec.Command("git", "-C", worktreePath, "config", "user.name", "Test").Run()
	exec.Command("git", "-C", worktreePath, "config", "user.email", "test@test.com").Run()

	// Make initial commit so we have a clean baseline
	f, _ := os.CreateTemp(worktreePath, "init-*.txt")
	f.WriteString("initial")
	f.Close()
	exec.Command("git", "-C", worktreePath, "add", ".").Run()
	exec.Command("git", "-C", worktreePath, "commit", "-m", "initial").Run()

	cleanup = func() {
		os.RemoveAll(tmpDir)
	}

	return worktreePath, cleanup
}

func TestCommitCount(t *testing.T) {
	bareRepo, cleanup := setupBareRepo(t)
	defer cleanup()

	tests := []struct {
		name   string
		branch string
		want   int
	}{
		{"main branch", "main", 0},
		{"feature branch", "feature-one", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CommitCount(bareRepo, tt.branch)
			if err != nil {
				t.Fatalf("CommitCount() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("CommitCount(%q) = %d, want %d", tt.branch, got, tt.want)
			}
		})
	}
}

func TestFirstCommitDate(t *testing.T) {
	bareRepo, cleanup := setupBareRepo(t)
	defer cleanup()

	date, err := FirstCommitDate(bareRepo, "main")
	if err != nil {
		t.Fatalf("FirstCommitDate() error: %v", err)
	}
	if date == "" {
		t.Fatal("FirstCommitDate() returned empty string")
	}
	// Should be a valid ISO-8601 datetime
	if !strings.Contains(date, "T") && !strings.Contains(date, "Z") &&
		!strings.Contains(date, "+") && !strings.Contains(date, "-") {
		t.Errorf("FirstCommitDate() = %q, expected ISO-8601 format", date)
	}

	// Feature branch first commit should be the same as main
	featureDate, err := FirstCommitDate(bareRepo, "feature-one")
	if err != nil {
		t.Fatalf("FirstCommitDate(feature-one) error: %v", err)
	}
	if featureDate == "" {
		t.Fatal("FirstCommitDate(feature-one) returned empty string")
	}
}

func TestLastCommitDate(t *testing.T) {
	bareRepo, cleanup := setupBareRepo(t)
	defer cleanup()

	// Main branch: last commit is the initial commit
	mainDate, err := LastCommitDate(bareRepo, "main")
	if err != nil {
		t.Fatalf("LastCommitDate(main) error: %v", err)
	}
	if mainDate == "" {
		t.Fatal("LastCommitDate(main) returned empty string")
	}
	if !strings.Contains(mainDate, "T") {
		t.Errorf("LastCommitDate(main) = %q, expected ISO-8601 format", mainDate)
	}

	// Feature branch last commit date should be later (or equal) to main
	featureDate, err := LastCommitDate(bareRepo, "feature-one")
	if err != nil {
		t.Fatalf("LastCommitDate(feature-one) error: %v", err)
	}
	if featureDate < mainDate {
		t.Errorf("feature last commit %q is before main last commit %q", featureDate, mainDate)
	}
}

func TestHasUncommittedChanges(t *testing.T) {
	worktreePath, cleanup := setupWorktree(t)
	defer cleanup()

	t.Run("clean repo", func(t *testing.T) {
		dirty, err := HasUncommittedChanges(worktreePath)
		if err != nil {
			t.Fatalf("HasUncommittedChanges() error: %v", err)
		}
		if dirty {
			t.Error("expected clean repo, got dirty")
		}
	})

	t.Run("dirty repo", func(t *testing.T) {
		// Create an untracked file
		f, err := os.CreateTemp(worktreePath, "dirty-*.txt")
		if err != nil {
			t.Fatalf("creating dirty file: %v", err)
		}
		f.WriteString("dirty")
		f.Close()

		dirty, err := HasUncommittedChanges(worktreePath)
		if err != nil {
			t.Fatalf("HasUncommittedChanges() error: %v", err)
		}
		if !dirty {
			t.Error("expected dirty repo, got clean")
		}
	})

	t.Run("modified file", func(t *testing.T) {
		// Modify a tracked file
		// First, get the list of existing files
		entries, err := os.ReadDir(worktreePath)
		if err != nil {
			t.Fatalf("readdir: %v", err)
		}
		var trackedFile string
		for _, e := range entries {
			if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				trackedFile = filepath.Join(worktreePath, e.Name())
				break
			}
		}
		if trackedFile == "" {
			t.Skip("no tracked files to modify")
		}

		// Append to the file
		f, err := os.OpenFile(trackedFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatalf("opening file: %v", err)
		}
		f.WriteString("\nmodification")
		f.Close()

		dirty, err := HasUncommittedChanges(worktreePath)
		if err != nil {
			t.Fatalf("HasUncommittedChanges() error: %v", err)
		}
		if !dirty {
			t.Error("expected dirty repo after modification")
		}
	})
}
