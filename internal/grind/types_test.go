package grind

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

var fixedTime = time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)

func ptr[T any](v T) *T { return &v }

func TestTotalSeconds(t *testing.T) {
	tests := []struct {
		name string
		time []Session
		want int64
	}{
		{"no sessions", nil, 0},
		{"empty sessions", []Session{}, 0},
		{"single session", []Session{{Rounded: 3600}}, 3600},
		{"multiple sessions", []Session{{Rounded: 3600}, {Rounded: 1800}, {Rounded: 900}}, 6300},
		{"sessions with zero", []Session{{Rounded: 0}, {Rounded: 3600}}, 3600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProjectConfig{Time: tt.time}
			if got := p.TotalSeconds(); got != tt.want {
				t.Errorf("TotalSeconds() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBilledSeconds(t *testing.T) {
	tru := true
	fal := false

	tests := []struct {
		name string
		time []Session
		want int64
	}{
		{"no sessions", nil, 0},
		{"none invoiced", []Session{{Rounded: 3600}, {Rounded: 1800}}, 0},
		{"all invoiced", []Session{{Rounded: 3600, Invoiced: &tru}, {Rounded: 1800, Invoiced: &tru}}, 5400},
		{"mixed invoiced", []Session{{Rounded: 3600, Invoiced: &tru}, {Rounded: 1800}, {Rounded: 900, Invoiced: &fal}}, 3600},
		{"explicit false", []Session{{Rounded: 3600, Invoiced: &fal}, {Rounded: 1800, Invoiced: &tru}}, 1800},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProjectConfig{Time: tt.time}
			if got := p.BilledSeconds(); got != tt.want {
				t.Errorf("BilledSeconds() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestUnbilledSeconds(t *testing.T) {
	tru := true

	tests := []struct {
		name string
		time []Session
		want int64
	}{
		{"no sessions", nil, 0},
		{"all unbilled", []Session{{Rounded: 3600}, {Rounded: 1800}}, 5400},
		{"some billed", []Session{{Rounded: 3600, Invoiced: &tru}, {Rounded: 1800}}, 1800},
		{"all billed", []Session{{Rounded: 3600, Invoiced: &tru}, {Rounded: 1800, Invoiced: &tru}}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProjectConfig{Time: tt.time}
			if got := p.UnbilledSeconds(); got != tt.want {
				t.Errorf("UnbilledSeconds() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestHours(t *testing.T) {
	tests := []struct {
		name          string
		time          []Session
		wantTotal     float64
		wantBilled    float64
		wantUnbilled  float64
		invoicedIndex int // index of session to mark invoiced, -1 for none
	}{
		{"no sessions", nil, 0, 0, 0, -1},
		{"one hour", []Session{{Rounded: 3600}}, 1.0, 0, 1.0, -1},
		{"one hour billed", []Session{{Rounded: 3600}}, 1.0, 1.0, 0, 0},
		{"mixed", []Session{{Rounded: 3600}, {Rounded: 1800}, {Rounded: 900}}, 1.75, 1.0, 0.75, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tru := true
			if tt.invoicedIndex >= 0 && tt.invoicedIndex < len(tt.time) {
				tt.time[tt.invoicedIndex].Invoiced = &tru
			}
			p := ProjectConfig{Time: tt.time}

			if got := p.TotalHours(); math.Abs(got-tt.wantTotal) > 0.0001 {
				t.Errorf("TotalHours() = %v, want %v", got, tt.wantTotal)
			}
			if got := p.BilledHours(); math.Abs(got-tt.wantBilled) > 0.0001 {
				t.Errorf("BilledHours() = %v, want %v", got, tt.wantBilled)
			}
			if got := p.UnbilledHours(); math.Abs(got-tt.wantUnbilled) > 0.0001 {
				t.Errorf("UnbilledHours() = %v, want %v", got, tt.wantUnbilled)
			}
		})
	}
}

func TestAmounts(t *testing.T) {
	tru := true

	tests := []struct {
		name         string
		time         []Session
		rate         float64
		wantTotal    float64
		wantBilled   float64
		wantUnbilled float64
	}{
		{
			name:         "no time",
			time:         nil,
			rate:         100,
			wantTotal:    0,
			wantBilled:   0,
			wantUnbilled: 0,
		},
		{
			name:         "2 hours at $150",
			time:         []Session{{Rounded: 7200}},
			rate:         150,
			wantTotal:    300,
			wantBilled:   0,
			wantUnbilled: 300,
		},
		{
			name:         "mixed billing",
			time:         []Session{{Rounded: 3600, Invoiced: &tru}, {Rounded: 1800}},
			rate:         100,
			wantTotal:    150.0,  // 1.5h * 100
			wantBilled:   100.0,  // 1h * 100
			wantUnbilled: 50.0,   // 0.5h * 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProjectConfig{
				Time:    tt.time,
				Billing: BillingConfig{Rate: tt.rate},
			}

			if got := p.TotalAmount(); math.Abs(got-tt.wantTotal) > 0.005 {
				t.Errorf("TotalAmount() = %v, want %v", got, tt.wantTotal)
			}
			if got := p.BilledAmount(); math.Abs(got-tt.wantBilled) > 0.005 {
				t.Errorf("BilledAmount() = %v, want %v", got, tt.wantBilled)
			}
			if got := p.UnbilledAmount(); math.Abs(got-tt.wantUnbilled) > 0.005 {
				t.Errorf("UnbilledAmount() = %v, want %v", got, tt.wantUnbilled)
			}
		})
	}
}

func TestActiveSession(t *testing.T) {
	now := fixedTime
	earlier := now.Add(-2 * time.Hour)

	tests := []struct {
		name string
		time []Session
		want *Session
	}{
		{"no sessions", nil, nil},
		{"no active session (all ended)", []Session{
			{Start: earlier, End: &now, Duration: 7200, Rounded: 7200},
		}, nil},
		{"has active session", []Session{
			{Start: earlier, End: &now, Duration: 7200, Rounded: 7200},
			{Start: now, End: nil, Duration: 0, Rounded: 0},
		}, &Session{Start: now, End: nil, Duration: 0, Rounded: 0}},
		{"only active session", []Session{
			{Start: now, End: nil, Duration: 0, Rounded: 0},
		}, &Session{Start: now, End: nil, Duration: 0, Rounded: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProjectConfig{Time: tt.time}
			got := p.ActiveSession()

			if tt.want == nil {
				if got != nil {
					t.Errorf("ActiveSession() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("ActiveSession() = nil, want %v", tt.want)
			}
			if !got.Start.Equal(tt.want.Start) {
				t.Errorf("ActiveSession().Start = %v, want %v", got.Start, tt.want.Start)
			}
			if got.End != nil || tt.want.End != nil {
				// Both nil is fine, check otherwise
			}
		})
	}
}

func TestLastSession(t *testing.T) {
	t1 := fixedTime
	t2 := t1.Add(1 * time.Hour)
	t3 := t2.Add(1 * time.Hour)

	tests := []struct {
		name string
		time []Session
		want *Session
	}{
		{"no sessions", nil, nil},
		{"single session", []Session{
			{Start: t1, Duration: 3600, Rounded: 3600},
		}, &Session{Start: t1, Duration: 3600, Rounded: 3600}},
		{"multiple sessions, last is latest", []Session{
			{Start: t1, Duration: 3600, Rounded: 3600},
			{Start: t3, Duration: 1800, Rounded: 1800},
			{Start: t2, Duration: 900, Rounded: 900},
		}, &Session{Start: t3, Duration: 1800, Rounded: 1800}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProjectConfig{Time: tt.time}
			got := p.LastSession()

			if tt.want == nil {
				if got != nil {
					t.Errorf("LastSession() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("LastSession() = nil, want %v", tt.want)
			}
			if !got.Start.Equal(tt.want.Start) {
				t.Errorf("LastSession().Start = %v, want %v", got.Start, tt.want.Start)
			}
		})
	}
}

func TestDurationHuman(t *testing.T) {
	tests := []struct {
		name     string
		duration int64
		want     string
	}{
		{"zero", 0, "0s"},
		{"seconds only", 45, "45s"},
		{"one minute", 60, "1m"},
		{"minutes only", 125, "2m"},
		{"one hour", 3600, "1h"},
		{"hours and minutes", 8100, "2h 15m"},
		{"exact hours", 7200, "2h"},
		{"complex", 3661, "1h 1m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Session{Duration: tt.duration}
			if got := s.DurationHuman(); got != tt.want {
				t.Errorf("DurationHuman() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJSONUnmarshal(t *testing.T) {
	// Verify that the types match the GrindCLI .project.json schema.
	input := `{
		"name": "test-project",
		"type": "app",
		"idea": "A test project",
		"billing": {
			"roundTo": "quarter-hour",
			"rate": 150.0
		},
		"repo": "git@github.com:user/repo.git",
		"code": "ABC-123",
		"longTerm": true,
		"time": [
			{
				"start": "2025-06-01T10:00:00Z",
				"end": "2025-06-01T12:30:00Z",
				"duration": 9000,
				"rounded": 9000,
				"invoiced": true
			}
		],
		"client": {
			"contact": "Jane Doe",
			"company": "Acme Corp",
			"email": "jane@acme.com"
		},
		"publications": [
			{
				"url": "https://example.com/post",
				"publishedAt": "2025-06-02"
			}
		]
	}`

	var p ProjectConfig
	if err := json.Unmarshal([]byte(input), &p); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if p.Name != "test-project" {
		t.Errorf("Name = %q, want %q", p.Name, "test-project")
	}
	if p.Billing.Rate != 150.0 {
		t.Errorf("Rate = %v, want %v", p.Billing.Rate, 150.0)
	}
	if len(p.Time) != 1 {
		t.Fatalf("len(Time) = %d, want 1", len(p.Time))
	}
	if p.Time[0].Duration != 9000 {
		t.Errorf("Duration = %d, want 9000", p.Time[0].Duration)
	}
	if p.Time[0].Invoiced == nil || !*p.Time[0].Invoiced {
		t.Error("Invoiced should be true")
	}
	if p.Client == nil {
		t.Fatal("Client should not be nil")
	}
	if p.Client.Company != "Acme Corp" {
		t.Errorf("Company = %q, want %q", p.Client.Company, "Acme Corp")
	}
	if len(p.Publications) != 1 {
		t.Fatalf("len(Publications) = %d, want 1", len(p.Publications))
	}
	if p.LongTerm != true {
		t.Error("LongTerm should be true")
	}
	if p.Code != "ABC-123" {
		t.Errorf("Code = %q, want %q", p.Code, "ABC-123")
	}
}
