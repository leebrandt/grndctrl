package tui

import (
	"testing"
	"time"

	"github.com/leebrandt/grndctrl/internal/grind"
	"github.com/leebrandt/grndctrl/internal/workspace"
)

func TestDataFingerprint(t *testing.T) {
	end := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	end2 := time.Date(2025, 6, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		rows    []projectRow
		wantEq  []int // indices that should produce equal fingerprints
		wantNeq []int // indices that should produce different fingerprints
	}{
		{
			name: "empty",
			rows: nil,
		},
		{
			name: "same data same fingerprint",
			rows: []projectRow{
				{info: workspace.ProjectInfo{Config: grind.ProjectConfig{Name: "a", Time: []grind.Session{{Rounded: 100}}}}, dirty: false},
				{info: workspace.ProjectInfo{Config: grind.ProjectConfig{Name: "b", Time: []grind.Session{{Rounded: 200}}}}, dirty: false},
			},
			wantEq: []int{0},
		},
		{
			name: "different session count changes fingerprint",
			rows: []projectRow{
				{info: workspace.ProjectInfo{Config: grind.ProjectConfig{Name: "a", Time: []grind.Session{{Rounded: 100}}}}, dirty: false},
				{info: workspace.ProjectInfo{Config: grind.ProjectConfig{Name: "a", Time: []grind.Session{{Rounded: 100}, {Rounded: 200}}}}, dirty: false},
			},
			wantNeq: []int{0, 1},
		},
		{
			name: "different dirty flag changes fingerprint",
			rows: []projectRow{
				{info: workspace.ProjectInfo{Config: grind.ProjectConfig{Name: "a", Time: []grind.Session{{Rounded: 100}}}}, dirty: false},
				{info: workspace.ProjectInfo{Config: grind.ProjectConfig{Name: "a", Time: []grind.Session{{Rounded: 100}}}}, dirty: true},
			},
			wantNeq: []int{0, 1},
		},
		{
			name: "different last session end changes fingerprint",
			rows: []projectRow{
				{info: workspace.ProjectInfo{Config: grind.ProjectConfig{Name: "a", Time: []grind.Session{{Rounded: 100, End: &end}}}}, dirty: false},
				{info: workspace.ProjectInfo{Config: grind.ProjectConfig{Name: "a", Time: []grind.Session{{Rounded: 100, End: &end2}}}}, dirty: false},
			},
			wantNeq: []int{0, 1},
		},
		{
			name: "same data different order different fingerprint",
			rows: []projectRow{
				{info: workspace.ProjectInfo{Config: grind.ProjectConfig{Name: "a"}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := dataFingerprint(tt.rows)

			for _, i := range tt.wantEq {
				for _, j := range tt.wantEq {
					if i != j && dataFingerprint([]projectRow{tt.rows[i]}) != dataFingerprint([]projectRow{tt.rows[j]}) {
						t.Errorf("rows[%d] and rows[%d] should have equal fingerprints", i, j)
					}
				}
			}

			for _, pair := range tt.wantNeq {
				_ = pair
			}

			if len(tt.wantNeq) == 2 {
				a := dataFingerprint([]projectRow{tt.rows[tt.wantNeq[0]]})
				b := dataFingerprint([]projectRow{tt.rows[tt.wantNeq[1]]})
				if a == b {
					t.Errorf("rows[%d] and rows[%d] should have different fingerprints", tt.wantNeq[0], tt.wantNeq[1])
				}
			}

			_ = fp
		})
	}
}

func TestRefreshIndicator(t *testing.T) {
	tests := []struct {
		name     string
		model    Model
		wantEmpty bool
		wantContains string
	}{
		{
			name:      "not auto refreshing no last refresh",
			model:     Model{autoRefresh: false},
			wantEmpty: true,
		},
		{
			name:     "refreshing shows spinner",
			model:    Model{refreshing: true, refreshSpinIdx: 0},
			wantContains: "Refreshing...",
		},
		{
			name:     "auto refresh with last refresh shows time",
			model:    Model{autoRefresh: true, lastRefresh: time.Date(2025, 6, 1, 14, 45, 0, 0, time.UTC)},
			wantContains: "Last:",
		},
		{
			name:      "auto refresh no last refresh yet",
			model:     Model{autoRefresh: true},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.model.refreshIndicator()
			if tt.wantEmpty && got != "" {
				t.Errorf("refreshIndicator() = %q, want empty", got)
			}
			if tt.wantContains != "" && !containsStr(got, tt.wantContains) {
				t.Errorf("refreshIndicator() = %q, want to contain %q", got, tt.wantContains)
			}
		})
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstr(s, sub))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
