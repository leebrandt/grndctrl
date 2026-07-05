// Package grind contains Go types and functions for reading Grind project data
// from .project.json files and querying git state.
package grind

import (
	"fmt"
	"math"
	"time"
)

// Session represents a single work session for a project.
type Session struct {
	Start    time.Time `json:"start"`
	End      *time.Time `json:"end,omitempty"`
	Duration int64     `json:"duration"`     // seconds
	Rounded  int64     `json:"rounded"`      // rounded seconds
	Invoiced *bool     `json:"invoiced,omitempty"`
}

// DurationHuman returns a human-readable duration string like "2h 15m".
func (s Session) DurationHuman() string {
	d := s.Duration
	hours := d / 3600
	minutes := (d % 3600) / 60
	secs := d % 60

	switch {
	case hours > 0 && minutes > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	case minutes > 0:
		return fmt.Sprintf("%dm", minutes)
	default:
		return fmt.Sprintf("%ds", secs)
	}
}

// BillingConfig describes how billing is rounded and the hourly rate.
type BillingConfig struct {
	RoundTo string  `json:"roundTo"` // "quarter-hour" | "half-hour" | "hour"
	Rate    float64 `json:"rate"`
}

// ClientInfo holds optional client contact details.
type ClientInfo struct {
	Contact string `json:"contact,omitempty"`
	Company string `json:"company,omitempty"`
	Address string `json:"address,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Email   string `json:"email,omitempty"`
}

// Publication represents a published article or post related to a project.
type Publication struct {
	URL         string `json:"url"`
	PublishedAt string `json:"publishedAt"`
}

// ProjectConfig is the top-level configuration for a Grind project, matching
// the GrindCLI .project.json schema.
type ProjectConfig struct {
	Name         string        `json:"name"`
	Type         string        `json:"type,omitempty"`
	Idea         string        `json:"idea"`
	Time         []Session     `json:"time"`
	Billing      BillingConfig `json:"billing"`
	Client       *ClientInfo   `json:"client,omitempty"`
	Repo         string        `json:"repo,omitempty"`
	Code         string        `json:"code,omitempty"`
	LongTerm     bool          `json:"longTerm,omitempty"`
	Publications []Publication `json:"publications,omitempty"`
}

// TotalSeconds returns the sum of all session rounded seconds.
func (p ProjectConfig) TotalSeconds() int64 {
	var total int64
	for _, s := range p.Time {
		total += s.Rounded
	}
	return total
}

// BilledSeconds returns the sum of all invoiced session rounded seconds.
func (p ProjectConfig) BilledSeconds() int64 {
	var total int64
	for _, s := range p.Time {
		if s.Invoiced != nil && *s.Invoiced {
			total += s.Rounded
		}
	}
	return total
}

// UnbilledSeconds returns TotalSeconds - BilledSeconds.
func (p ProjectConfig) UnbilledSeconds() int64 {
	return p.TotalSeconds() - p.BilledSeconds()
}

// TotalHours returns total hours as a decimal.
func (p ProjectConfig) TotalHours() float64 {
	return float64(p.TotalSeconds()) / 3600.0
}

// BilledHours returns billed hours as a decimal.
func (p ProjectConfig) BilledHours() float64 {
	return float64(p.BilledSeconds()) / 3600.0
}

// UnbilledHours returns unbilled hours as a decimal.
func (p ProjectConfig) UnbilledHours() float64 {
	return float64(p.UnbilledSeconds()) / 3600.0
}

// TotalAmount returns the total dollar amount (hours * rate), rounded to cents.
func (p ProjectConfig) TotalAmount() float64 {
	return math.Round(p.TotalHours()*p.Billing.Rate*100) / 100
}

// BilledAmount returns the billed dollar amount, rounded to cents.
func (p ProjectConfig) BilledAmount() float64 {
	return math.Round(p.BilledHours()*p.Billing.Rate*100) / 100
}

// UnbilledAmount returns the unbilled dollar amount, rounded to cents.
func (p ProjectConfig) UnbilledAmount() float64 {
	return math.Round(p.UnbilledHours()*p.Billing.Rate*100) / 100
}

// ActiveSession returns the session with no end time, or nil if none.
func (p ProjectConfig) ActiveSession() *Session {
	for i := range p.Time {
		if p.Time[i].End == nil {
			return &p.Time[i]
		}
	}
	return nil
}

// LastSession returns the most recent session by Start time, or nil.
func (p ProjectConfig) LastSession() *Session {
	if len(p.Time) == 0 {
		return nil
	}
	last := p.Time[0]
	for _, s := range p.Time[1:] {
		if s.Start.After(last.Start) {
			last = s
		}
	}
	return &last
}
