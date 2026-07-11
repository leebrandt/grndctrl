package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func setupIdeasFixture(t *testing.T) (workspaceRoot string, cleanup func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "grind-ideas-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}

	ideasDir := filepath.Join(tmpDir, "grind", "ideas")
	if err := os.MkdirAll(ideasDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("creating ideas dir: %v", err)
	}

	files := map[string]string{
		"20260125051508.md": "# The Misfortune of Meaning\n\nA deep dive into existential dread.\n",
		"20260203100000.md": "# AI Dotfiles\n\nUsing AI to manage dotfiles.\n",
		"20260115120000.md": "# Bad Idea\n\nThis one is terrible.\n",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(ideasDir, name), []byte(content), 0644); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("writing %s: %v", name, err)
		}
	}

	rejectedDir := filepath.Join(tmpDir, "grind", "ideas")
	rejectedFile := filepath.Join(rejectedDir, "rejected-20260110080000.md")
	if err := os.WriteFile(rejectedFile, []byte("# Terrible Plan\n\nNo good at all.\n"), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("writing rejected file: %v", err)
	}

	cleanup = func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestCollectIdeas(t *testing.T) {
	workspaceRoot, cleanup := setupIdeasFixture(t)
	defer cleanup()

	t.Run("default excludes rejected", func(t *testing.T) {
		ideas, err := CollectIdeas(workspaceRoot, false)
		if err != nil {
			t.Fatalf("CollectIdeas() error: %v", err)
		}
		if len(ideas) != 3 {
			t.Fatalf("got %d ideas, want 3", len(ideas))
		}
		for _, idea := range ideas {
			if idea.Rejected {
				t.Errorf("idea %q should not be in default results", idea.Filename)
			}
		}
	})

	t.Run("include rejected", func(t *testing.T) {
		ideas, err := CollectIdeas(workspaceRoot, true)
		if err != nil {
			t.Fatalf("CollectIdeas() error: %v", err)
		}
		if len(ideas) != 4 {
			t.Fatalf("got %d ideas, want 4", len(ideas))
		}

		rejectedCount := 0
		for _, idea := range ideas {
			if idea.Rejected {
				rejectedCount++
			}
		}
		if rejectedCount != 1 {
			t.Errorf("got %d rejected, want 1", rejectedCount)
		}
	})

	t.Run("sorted by created time", func(t *testing.T) {
		ideas, err := CollectIdeas(workspaceRoot, true)
		if err != nil {
			t.Fatalf("CollectIdeas() error: %v", err)
		}
		for i := 1; i < len(ideas); i++ {
			if ideas[i].Created.Before(ideas[i-1].Created) {
				t.Errorf("ideas not sorted: %v before %v", ideas[i].Created, ideas[i-1].Created)
			}
		}
	})

	t.Run("numbers are sequential", func(t *testing.T) {
		ideas, err := CollectIdeas(workspaceRoot, false)
		if err != nil {
			t.Fatalf("CollectIdeas() error: %v", err)
		}
		for i, idea := range ideas {
			if idea.Number != i {
				t.Errorf("idea[%d].Number = %d", i, idea.Number)
			}
		}
	})

	t.Run("titles parsed correctly", func(t *testing.T) {
		ideas, err := CollectIdeas(workspaceRoot, false)
		if err != nil {
			t.Fatalf("CollectIdeas() error: %v", err)
		}
		titles := make(map[string]string)
		for _, idea := range ideas {
			titles[idea.Filename] = idea.Title
		}
		if titles["20260125051508.md"] != "The Misfortune of Meaning" {
			t.Errorf("title = %q, want %q", titles["20260125051508.md"], "The Misfortune of Meaning")
		}
		if titles["20260203100000.md"] != "AI Dotfiles" {
			t.Errorf("title = %q, want %q", titles["20260203100000.md"], "AI Dotfiles")
		}
	})

	t.Run("empty ideas directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		ideasDir := filepath.Join(tmpDir, "grind", "ideas")
		os.MkdirAll(ideasDir, 0755)

		ideas, err := CollectIdeas(tmpDir, false)
		if err != nil {
			t.Fatalf("CollectIdeas() error: %v", err)
		}
		if len(ideas) != 0 {
			t.Fatalf("got %d ideas, want 0", len(ideas))
		}
	})

	t.Run("missing ideas directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		ideas, err := CollectIdeas(tmpDir, false)
		if err != nil {
			t.Fatalf("CollectIdeas() error: %v", err)
		}
		if len(ideas) != 0 {
			t.Fatalf("got %d ideas, want 0", len(ideas))
		}
	})

	t.Run("untitled idea", func(t *testing.T) {
		tmpDir := t.TempDir()
		ideasDir := filepath.Join(tmpDir, "grind", "ideas")
		os.MkdirAll(ideasDir, 0755)

		emptyFile := filepath.Join(ideasDir, "20260301120000.md")
		if err := os.WriteFile(emptyFile, []byte("Just some text, no heading.\n"), 0644); err != nil {
			t.Fatalf("writing file: %v", err)
		}

		ideas, err := CollectIdeas(tmpDir, false)
		if err != nil {
			t.Fatalf("CollectIdeas() error: %v", err)
		}
		if len(ideas) != 1 {
			t.Fatalf("got %d ideas, want 1", len(ideas))
		}
		if ideas[0].Title != "(untitled)" {
			t.Errorf("title = %q, want %q", ideas[0].Title, "(untitled)")
		}
	})
}

func TestParseIdeaTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"valid", "20260125051508.md", false},
		{"rejected", "rejected-20260125051508.md", false},
		{"wrong length", "20260125.md", true},
		{"non-numeric", "abcdefghijklmnop.md", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseIdeaTimestamp(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIdeaTimestamp(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			}
		})
	}
}
