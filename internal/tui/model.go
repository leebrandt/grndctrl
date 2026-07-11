package tui

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/leebrandt/grndctrl/internal/grind"
	"github.com/leebrandt/grndctrl/internal/workspace"
)

type viewName int

const (
	viewDashboard viewName = iota
	viewIdeas
	viewDetail
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

	// scroll tracking
	scrollOffset int

	// filter mode
	filterMode  bool
	filterText  string
	filtered    []int // indices into m.projects matching filter

	// help overlay
	showHelp bool

	// view / focus
	currentView viewName

	// detail view
	detailProject  int // index into m.projects for the detail view
	detailViewport viewport.Model
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
		workspace:      ws,
		loading:        true,
		spinner:        s,
		currentView:    viewDashboard,
		detailViewport: viewport.New(80, 20),
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
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.ready = true
		m.width = msg.Width
		m.height = msg.Height
		m.detailViewport = viewport.New(msg.Width-4, m.detailContentHeight())
		m.detailViewport.Style = lipgloss.NewStyle().Padding(0, 2)
		return m, nil

	case ProjectsLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.projects = msg.Projects
		m.activeSessions = collectActiveSessions(msg.Projects)
		m.rebuildFilter()
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

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys (work regardless of mode)
	switch msg.String() {
	case "q", "ctrl+c":
		return m.quit()
	}

	// Detail view keys
	if m.currentView == viewDetail {
		return m.handleDetailKey(msg)
	}

	// Filter mode keys
	if m.filterMode {
		return m.handleFilterKey(msg)
	}

	// Help overlay keys
	if m.showHelp {
		if msg.String() == "?" || msg.String() == "esc" {
			m.showHelp = false
		}
		return m, nil
	}

	// Normal mode keys
	switch msg.String() {
	case "esc":
		return m.quit()
	case "j", "down":
		m.cursorDown()
	case "k", "up":
		m.cursorUp()
	case "g":
		m.cursorToTop()
	case "G":
		m.cursorToBottom()
	case "ctrl+d":
		m.pageDown()
	case "ctrl+u":
		m.pageUp()
	case "/":
		m.filterMode = true
		m.filterText = ""
	case "?":
		m.showHelp = !m.showHelp
	case "r":
		return m.refresh()
	case "i":
		m.currentView = viewIdeas
	case "tab":
		m.toggleFocus()
	case "enter":
		m.openDetail()
	}

	return m, nil
}

// handleDetailKey handles key events when the detail view is active.
func (m *Model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.currentView = viewDashboard
		return m, nil
	case "j", "down":
		m.detailViewport.LineDown(1)
	case "k", "up":
		m.detailViewport.LineUp(1)
	case "g":
		m.detailViewport.SetYOffset(0)
	case "G":
		m.detailViewport.SetYOffset(m.detailViewport.YOffset + 9999)
	case "ctrl+d":
		m.detailViewport.HalfViewDown()
	case "ctrl+u":
		m.detailViewport.HalfViewUp()
	}
	return m, nil
}

// openDetail transitions to the detail view for the currently selected project.
func (m *Model) openDetail() {
	if len(m.projects) == 0 {
		return
	}
	m.detailProject = m.cursor
	m.currentView = viewDetail
	m.detailViewport.SetYOffset(0)
	m.detailViewport.SetContent(m.detailView())
}

// detailContentHeight returns the available height for the detail viewport.
func (m *Model) detailContentHeight() int {
	bannerLines := len(m.activeSessions)
	if bannerLines > 0 {
		bannerLines++ // separator
	}
	// 1 for status bar
	overhead := 1 + bannerLines
	ch := m.height - overhead
	if ch < 1 {
		ch = 1
	}
	return ch
}

func (m *Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filterMode = false
		m.filterText = ""
		m.rebuildFilter()
	case "enter":
		m.filterMode = false
		// Filter stays applied
	case "backspace":
		if len(m.filterText) > 0 {
			m.filterText = m.filterText[:len(m.filterText)-1]
			m.rebuildFilter()
		}
	default:
		// Accept printable characters
		if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] <= 126 {
			m.filterText += msg.String()
			m.rebuildFilter()
		}
	}
	return m, nil
}

func (m *Model) rebuildFilter() {
	if m.filterText == "" {
		m.filtered = nil
		return
	}
	lower := strings.ToLower(m.filterText)
	m.filtered = nil
	for i, p := range m.projects {
		name := strings.ToLower(p.info.Config.Name)
		typ := strings.ToLower(p.info.Config.Type)
		if strings.Contains(name, lower) || strings.Contains(typ, lower) {
			m.filtered = append(m.filtered, i)
		}
	}
	// Clamp cursor to filtered range
	if len(m.filtered) > 0 && m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
}

// visibleProjects returns the slice of projects to display (filtered or all).
func (m *Model) visibleProjects() []projectRow {
	if m.filtered != nil {
		visible := make([]projectRow, len(m.filtered))
		for i, idx := range m.filtered {
			visible[i] = m.projects[idx]
		}
		return visible
	}
	return m.projects
}

// visibleCursor returns the cursor position in the visible list.
func (m *Model) visibleCursor() int {
	if m.filtered != nil {
		for i, idx := range m.filtered {
			if idx == m.cursor {
				return i
			}
		}
		return 0
	}
	return m.cursor
}

// visibleCount returns the number of visible projects.
func (m *Model) visibleCount() int {
	if m.filtered != nil {
		return len(m.filtered)
	}
	return len(m.projects)
}

func (m *Model) cursorDown() {
	count := m.visibleCount()
	if count == 0 {
		return
	}
	vc := m.visibleCursor()
	if vc < count-1 {
		// Advance the real cursor to the next visible project
		if m.filtered != nil {
			m.cursor = m.filtered[vc+1]
		} else {
			m.cursor++
		}
		m.clampScrollOffset()
	}
}

func (m *Model) cursorUp() {
	count := m.visibleCount()
	if count == 0 {
		return
	}
	vc := m.visibleCursor()
	if vc > 0 {
		if m.filtered != nil {
			m.cursor = m.filtered[vc-1]
		} else {
			m.cursor--
		}
		m.clampScrollOffset()
	}
}

func (m *Model) cursorToTop() {
	if m.visibleCount() == 0 {
		return
	}
	if m.filtered != nil {
		m.cursor = m.filtered[0]
	} else {
		m.cursor = 0
	}
	m.scrollOffset = 0
}

func (m *Model) cursorToBottom() {
	count := m.visibleCount()
	if count == 0 {
		return
	}
	if m.filtered != nil {
		m.cursor = m.filtered[count-1]
	} else {
		m.cursor = count - 1
	}
	m.clampScrollOffset()
}

func (m *Model) pageDown() {
	pageSize := m.pageSize()
	vc := m.visibleCursor()
	target := vc + pageSize
	count := m.visibleCount()
	if target >= count {
		target = count - 1
	}
	if m.filtered != nil {
		m.cursor = m.filtered[target]
	} else {
		m.cursor = target
	}
	m.clampScrollOffset()
}

func (m *Model) pageUp() {
	pageSize := m.pageSize()
	vc := m.visibleCursor()
	target := vc - pageSize
	if target < 0 {
		target = 0
	}
	if m.filtered != nil {
		m.cursor = m.filtered[target]
	} else {
		m.cursor = target
	}
	m.clampScrollOffset()
}

// pageSize returns half the available content height.
func (m *Model) pageSize() int {
	contentHeight := m.contentHeight()
	if contentHeight < 2 {
		return 1
	}
	return contentHeight / 2
}

// contentHeight returns the number of rows available for the project table.
func (m *Model) contentHeight() int {
	bannerLines := len(m.activeSessions)
	if bannerLines > 0 {
		bannerLines++ // separator
	}
	// 1 for title, 1 for header, 1 for separator, 1 for status bar, 2 for spacing
	overhead := 6 + bannerLines
	ch := m.height - overhead
	if ch < 1 {
		ch = 1
	}
	return ch
}

// clampScrollOffset ensures the cursor is visible in the viewport.
func (m *Model) clampScrollOffset() {
	ch := m.contentHeight()
	vc := m.visibleCursor()

	if vc < m.scrollOffset {
		m.scrollOffset = vc
	}
	if vc >= m.scrollOffset+ch {
		m.scrollOffset = vc - ch + 1
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m *Model) toggleFocus() {
	if m.currentView == viewDashboard {
		m.currentView = viewIdeas
	} else {
		m.currentView = viewDashboard
	}
}

func (m *Model) quit() (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m Model) refresh() (tea.Model, tea.Cmd) {
	return m, loadProjects(m.workspace)
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

	var content string
	if m.currentView == viewDetail {
		content = m.detailViewport.View()
	} else {
		content = m.dashboardView()
	}

	if banner != "" {
		content = banner + "\n" + content
	}

	// Help overlay on top of everything
	if m.showHelp {
		content = m.helpOverlay(content)
	}

	// Filter prompt at bottom
	if m.filterMode {
		content = m.filterPrompt(content)
	}

	return content
}

func (m Model) filterPrompt(viewContent string) string {
	promptText := "Filter: " + m.filterText + "▌"
	prompt := FilterPromptStyle.Render(promptText)

	// Place the prompt at the bottom of the view
	lines := strings.Split(viewContent, "\n")
	if len(lines) > 0 {
		lines[len(lines)-1] = prompt
	}
	return strings.Join(lines, "\n")
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
