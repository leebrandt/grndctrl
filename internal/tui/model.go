package tui

import (
	"path/filepath"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/leebrandt/grndctrl/internal/grind"
	"github.com/leebrandt/grndctrl/internal/workspace"
)

type projectRow struct {
	info           workspace.ProjectInfo
	dirty          bool
	lastCommitDate string
	gitErr         bool
}

type activeSessionInfo struct {
	projectName string
	start       time.Time
	rate        float64
	roundTo     string
}

type Model struct {
	ready     bool
	width     int
	height    int
	workspace string
	projects  []projectRow
	cursor    int
	loading   bool
	err       error
	spinner   spinner.Model

	activeSessions []activeSessionInfo
	tickPulse      bool
}

type ProjectsLoadedMsg struct {
	Projects []projectRow
	Err      error
}

type tickMsg time.Time

func NewModel(ws string) Model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent))
	s.Spinner = spinner.MiniDot

	return Model{
		workspace: ws,
		loading:   true,
		spinner:   s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadProjects(m.workspace),
		m.spinner.Tick,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if !m.loading && len(m.projects) > 0 {
				if m.cursor < len(m.projects)-1 {
					m.cursor++
				}
			}
			return m, nil
		case "k", "up":
			if !m.loading && len(m.projects) > 0 {
				if m.cursor > 0 {
					m.cursor--
				}
			}
			return m, nil
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.ready = true
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case ProjectsLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.projects = msg.Projects
		m.activeSessions = collectActiveSessions(msg.Projects)
		cmds := []tea.Cmd{m.activeSessionTick()}
		return m, tea.Batch(cmds...)

	case tickMsg:
		m.tickPulse = !m.tickPulse
		m.activeSessions = collectActiveSessions(m.projects)
		if len(m.activeSessions) > 0 {
			return m, m.activeSessionTick()
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

func collectActiveSessions(rows []projectRow) []activeSessionInfo {
	var sessions []activeSessionInfo
	for _, row := range rows {
		s := row.info.Config.ActiveSession()
		if s != nil {
			sessions = append(sessions, activeSessionInfo{
				projectName: row.info.Config.Name,
				start:       s.Start,
				rate:        row.info.Config.Billing.Rate,
				roundTo:     row.info.Config.Billing.RoundTo,
			})
		}
	}
	return sessions
}

func (m Model) activeSessionTick() tea.Cmd {
	if len(m.activeSessions) == 0 {
		return nil
	}
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	if m.loading {
		return m.loadingView()
	}

	if m.err != nil {
		return m.errorView()
	}

	if len(m.projects) == 0 {
		return m.emptyView()
	}

	banner := m.activeSessionBanner()
	content := m.dashboardView()

	if banner != "" {
		content = banner + "\n" + content
	}

	return content
}

func (m Model) loadingView() string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		TitleStyle.Render("GRNDCTRL"),
		"",
		DimStyle.Render(m.spinner.View()+" Loading projects..."),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) errorView() string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		ErrorStyle.Render("Error loading projects:"),
		"",
		DimStyle.Render(m.err.Error()),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) emptyView() string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		TitleStyle.Render("GRNDCTRL"),
		"",
		DimStyle.Render("No active projects."),
		DimStyle.Render(`Create one with "grind new project"`),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func loadProjects(ws string) tea.Cmd {
	return func() tea.Msg {
		infos, err := workspace.CollectProjectInfos(ws)
		if err != nil {
			return ProjectsLoadedMsg{Err: err}
		}

		bareRepo := filepath.Join(ws, ".grind.repo.git")

		rows := make([]projectRow, 0, len(infos))
		for _, info := range infos {
			row := projectRow{info: info}

			dirty, err := grind.HasUncommittedChanges(info.WorktreePath)
			if err == nil {
				row.dirty = dirty
			} else {
				row.gitErr = true
			}

			date, err := grind.LastCommitDate(bareRepo, info.Branch)
			if err == nil {
				row.lastCommitDate = date
			} else {
				row.gitErr = true
			}

			rows = append(rows, row)
		}

		sortProjectRows(rows)

		return ProjectsLoadedMsg{Projects: rows}
	}
}

func sortProjectRows(rows []projectRow) {
	sort.Slice(rows, func(i, j int) bool {
		iLast := rows[i].info.Config.LastSession()
		jLast := rows[j].info.Config.LastSession()

		switch {
		case iLast == nil && jLast == nil:
			return rows[i].info.Config.Name < rows[j].info.Config.Name
		case iLast == nil:
			return true
		case jLast == nil:
			return false
		default:
			return iLast.Start.Before(jLast.Start)
		}
	})
}
