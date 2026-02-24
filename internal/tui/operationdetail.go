package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sanity-labs/blueprints-tui/internal/api"
)

const operationHeaderChrome = 4 // timestamp + blank line + "Logs" header with bottom border

type operationLogsLoadedMsg struct {
	logs []api.Log
}

type operationDetailModel struct {
	operation   api.Operation
	client      *api.Client
	stackID     string
	logs        []api.Log
	viewport    viewport.Model
	spinner     spinner.Model
	loadingLogs bool
	err         error
}

func newOperationDetailModel(client *api.Client, stackID string, op api.Operation, width, height int) operationDetailModel {
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	vp := viewport.New(width, height-operationHeaderChrome)

	return operationDetailModel{
		operation:   op,
		client:      client,
		stackID:     stackID,
		viewport:    vp,
		spinner:     sp,
		loadingLogs: true,
	}
}

func (m *operationDetailModel) SetSize(w, h int) {
	m.viewport.Width = w
	m.viewport.Height = h - operationHeaderChrome
}

func (m operationDetailModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchLogs())
}

func (m operationDetailModel) Update(msg tea.Msg) (operationDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case operationLogsLoadedMsg:
		m.loadingLogs = false
		m.logs = msg.logs
		m.viewport.SetContent(formatLogs(msg.logs))
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

func (m operationDetailModel) View() string {
	var b strings.Builder

	b.WriteString(mutedStyle.Render(fmt.Sprintf("Created: %s", m.operation.CreatedAt.Format("2006-01-02 15:04:05"))))
	if m.operation.CompletedAt != nil {
		b.WriteString("  ")
		b.WriteString(mutedStyle.Render(fmt.Sprintf("Completed: %s", m.operation.CompletedAt.Format("2006-01-02 15:04:05"))))
	}
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render("Logs") + "\n")

	if m.err != nil {
		b.WriteString("Error: " + m.err.Error())
		return b.String()
	}
	if m.loadingLogs {
		b.WriteString(m.spinner.View() + " Loading logs…")
		return b.String()
	}

	b.WriteString(m.viewport.View())
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
