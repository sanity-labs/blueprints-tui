package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sanity-labs/blueprints-tui/internal/api"
)

type operationLogsLoadedMsg struct {
	logs []api.Log
}

type operationDetailModel struct {
	operation   api.Operation
	client      *api.Client
	stackID     string
	styles      styles
	logs        []api.Log
	viewport    viewport.Model
	spinner     spinner.Model
	loadingLogs bool
	err         error
	height      int
}

func newOperationDetailModel(client *api.Client, stackID string, op api.Operation, s styles, width, height int) operationDetailModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := operationDetailModel{
		operation:   op,
		client:      client,
		stackID:     stackID,
		styles:      s,
		spinner:     sp,
		loadingLogs: true,
		height:      height,
	}

	innerH := height - m.chromeHeight()
	if innerH < 1 {
		innerH = 1
	}
	m.viewport = viewport.New(viewport.WithWidth(width), viewport.WithHeight(innerH))

	return m
}

// chromeHeight returns the measured height of the fixed header region
// (operation info + "Logs" section heading).
func (m operationDetailModel) chromeHeight() int {
	return lipgloss.Height(m.renderChrome())
}

func (m *operationDetailModel) SetSize(w, h int) {
	m.height = h
	innerH := h - m.chromeHeight()
	if innerH < 1 {
		innerH = 1
	}
	m.viewport.SetWidth(w)
	m.viewport.SetHeight(innerH)
}

func (m operationDetailModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchLogs())
}

func (m operationDetailModel) Update(msg tea.Msg) (operationDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case operationLogsLoadedMsg:
		m.loadingLogs = false
		m.logs = msg.logs
		m.viewport.SetContent(m.formatLogs(msg.logs))
		m.viewport.GotoBottom()

	case apiErrMsg:
		m.loadingLogs = false
		m.err = msg.err

	case spinner.TickMsg:
		if m.loadingLogs {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View returns exactly m.height lines. Chrome is fixed; the inner area
// (viewport or spinner) fills the remaining space.
func (m operationDetailModel) View() string {
	chrome := m.renderChrome()

	var inner string
	if m.err != nil {
		inner = "Error: " + m.err.Error()
	} else if m.loadingLogs {
		inner = m.spinner.View() + " Loading logs…"
	} else if len(m.logs) == 0 {
		inner = m.styles.muted.Render("No logs available.")
	} else {
		inner = m.viewport.View()
	}

	innerH := m.height - m.chromeHeight()
	if innerH < 1 {
		innerH = 1
	}
	inner = lipgloss.PlaceVertical(innerH, lipgloss.Top, inner)

	return chrome + "\n" + inner
}

// renderChrome returns the operation header + "Logs" section heading.
func (m operationDetailModel) renderChrome() string {
	s := m.styles

	indicator := s.statusIndicator(m.operation.Status)
	line1 := fmt.Sprintf("Operation: %s  %s %s",
		m.operation.ID,
		indicator,
		s.statusStyle(m.operation.Status).Render(m.operation.Status),
	)

	var meta []string
	if m.operation.CompletedAt != nil {
		dur := m.operation.CompletedAt.Sub(m.operation.CreatedAt)
		meta = append(meta, fmt.Sprintf("%ds", int(dur.Seconds())))
	}
	meta = append(meta, "Created: "+m.operation.CreatedAt.Format("2006-01-02 15:04"))
	if m.operation.CompletedAt != nil {
		meta = append(meta, "Completed: "+m.operation.CompletedAt.Format("2006-01-02 15:04"))
	}
	line2 := s.muted.Render(strings.Join(meta, "  ·  "))

	logsHead := s.sectionHead.Render("Logs")

	return line1 + "\n" + line2 + "\n\n" + logsHead
}

func (m operationDetailModel) formatLogs(logs []api.Log) string {
	s := m.styles
	if len(logs) == 0 {
		return s.muted.Render("No logs available.")
	}
	var b strings.Builder
	for i := len(logs) - 1; i >= 0; i-- {
		l := logs[i]
		ts := s.muted.Render(l.Timestamp.Format("2006-01-02 15:04:05"))
		level := l.Level
		if level == "" {
			level = "INFO"
		}
		levelStr := s.logLevelStyle(level).Render(fmt.Sprintf("%-5s", level))
		b.WriteString(fmt.Sprintf("%s %s %s\n", ts, levelStr, l.Message))
	}
	return b.String()
}

func (m operationDetailModel) fetchLogs() tea.Cmd {
	return func() tea.Msg {
		logs, err := m.client.ListLogs(api.ListLogsOpts{
			OperationID: m.operation.ID,
		})
		if err != nil {
			return apiErrMsg{err: err}
		}
		return operationLogsLoadedMsg{logs: logs}
	}
}
