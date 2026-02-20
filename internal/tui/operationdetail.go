package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	"github.com/sanity-io/blueprints-tui/internal/api"
)

const operationChrome = 7 // breadcrumb + timestamps + header + spacing

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
	width       int
	height      int
}

func newOperationDetailModel(client *api.Client, stackID string, op api.Operation, width, height int) operationDetailModel {
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height-operationChrome))

	return operationDetailModel{
		operation:   op,
		client:      client,
		stackID:     stackID,
		viewport:    vp,
		spinner:     sp,
		width:       width,
		height:      height,
		loadingLogs: true,
	}
}

func (m operationDetailModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchLogs())
}

func (m operationDetailModel) Update(msg tea.Msg) (operationDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - operationChrome)

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
