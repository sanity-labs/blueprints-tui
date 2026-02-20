package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/sanity-io/blueprints-tui/internal/api"
)

type detailTab int

const (
	tabResources detailTab = iota
	tabOperations
	tabLogs
	tabCount
)

func (t detailTab) String() string {
	switch t {
	case tabResources:
		return "Resources"
	case tabOperations:
		return "Operations"
	case tabLogs:
		return "Logs"
	}
	return ""
}

const detailChrome = 8 // breadcrumb + tabs + help bar + spacing

type resourcesLoadedMsg struct {
	resources []api.Resource
}

type operationsLoadedMsg struct {
	operations []api.Operation
}

type logsLoadedMsg struct {
	logs []api.Log
}

type stackDetailModel struct {
	stack      api.Stack
	client     *api.Client
	activeTab  detailTab
	resources  []api.Resource
	operations []api.Operation
	logs       []api.Log

	resourceTable  table.Model
	operationTable table.Model
	logViewport    viewport.Model
	spinner        spinner.Model
	help           help.Model

	loadingResources  bool
	loadingOperations bool
	loadingLogs       bool
	err               error
	width             int
	height            int
}

func newStackDetailModel(client *api.Client, stack api.Stack, width, height int) stackDetailModel {
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	contentHeight := height - detailChrome

	rt := table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 30},
			{Title: "Type", Width: 30},
			{Title: "ID", Width: 16},
		}),
		table.WithFocused(true),
		table.WithWidth(width),
		table.WithHeight(contentHeight),
	)

	ot := table.New(
		table.WithColumns([]table.Column{
			{Title: "ID", Width: 16},
			{Title: "Status", Width: 14},
			{Title: "Created", Width: 20},
		}),
		table.WithWidth(width),
		table.WithHeight(contentHeight),
	)

	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(contentHeight))

	return stackDetailModel{
		stack:             stack,
		client:            client,
		activeTab:         tabResources,
		resourceTable:     rt,
		operationTable:    ot,
		logViewport:       vp,
		spinner:           sp,
		help:              help.New(),
		width:             width,
		height:            height,
		loadingResources:  true,
		loadingOperations: true,
		loadingLogs:       true,
	}
}

func (m stackDetailModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchResources(),
		m.fetchOperations(),
		m.fetchLogs(),
	)
}

func (m stackDetailModel) Update(msg tea.Msg) (stackDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.SetWidth(msg.Width)
		contentHeight := msg.Height - detailChrome
		m.resourceTable.SetWidth(msg.Width)
		m.resourceTable.SetHeight(contentHeight)
		m.operationTable.SetWidth(msg.Width)
		m.operationTable.SetHeight(contentHeight)
		m.logViewport.SetWidth(msg.Width)
		m.logViewport.SetHeight(contentHeight)

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, appKeys.Tab):
			m.activeTab = (m.activeTab + 1) % tabCount
			m = m.updateFocus()
		case key.Matches(msg, appKeys.ShiftTab):
			m.activeTab = (m.activeTab + tabCount - 1) % tabCount
			m = m.updateFocus()
		case key.Matches(msg, appKeys.Refresh):
			m.loadingResources = true
			m.loadingOperations = true
			m.loadingLogs = true
			return m, tea.Batch(m.fetchResources(), m.fetchOperations(), m.fetchLogs())
		}

	case resourcesLoadedMsg:
		m.loadingResources = false
		m.resources = msg.resources
		rows := make([]table.Row, len(msg.resources))
		for i, r := range msg.resources {
			rows[i] = table.Row{r.Name, r.Type, r.ID}
		}
		m.resourceTable.SetRows(rows)

	case operationsLoadedMsg:
		m.loadingOperations = false
		m.operations = msg.operations
		rows := make([]table.Row, len(msg.operations))
		for i, op := range msg.operations {
			rows[i] = table.Row{op.ID, op.Status, op.CreatedAt.Format("2006-01-02 15:04:05")}
		}
		m.operationTable.SetRows(rows)

	case logsLoadedMsg:
		m.loadingLogs = false
		m.logs = msg.logs
		m.logViewport.SetContent(formatLogs(msg.logs))
		m.logViewport.GotoBottom()

	case apiErrMsg:
		m.err = msg.err
		m.loadingResources = false
		m.loadingOperations = false
		m.loadingLogs = false

	case spinner.TickMsg:
		if m.isLoading() {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	switch m.activeTab {
	case tabResources:
		m.resourceTable, cmd = m.resourceTable.Update(msg)
	case tabOperations:
		m.operationTable, cmd = m.operationTable.Update(msg)
	case tabLogs:
		m.logViewport, cmd = m.logViewport.Update(msg)
	}
	return m, cmd
}

func (m stackDetailModel) View() string {
	var b strings.Builder

	b.WriteString(m.renderTabs() + "\n\n")

	if m.err != nil {
		b.WriteString("Error: " + m.err.Error())
		return b.String()
	}

	switch m.activeTab {
	case tabResources:
		if m.loadingResources {
			b.WriteString(m.spinner.View() + " Loading resources…")
		} else {
			b.WriteString(m.resourceTable.View())
		}
	case tabOperations:
		if m.loadingOperations {
			b.WriteString(m.spinner.View() + " Loading operations…")
		} else {
			b.WriteString(m.operationTable.View())
		}
	case tabLogs:
		if m.loadingLogs {
			b.WriteString(m.spinner.View() + " Loading logs…")
		} else {
			b.WriteString(m.logViewport.View())
		}
	}

	helpBar := helpStyle.Render(m.help.View(m))
	contentHeight := lipgloss.Height(b.String())
	helpHeight := lipgloss.Height(helpBar)
	if gap := m.height - contentHeight - helpHeight; gap > 0 {
		b.WriteString(strings.Repeat("\n", gap))
	}
	b.WriteString(helpBar)

	return b.String()
}

func (m stackDetailModel) renderTabs() string {
	tabs := make([]string, tabCount)
	for i := detailTab(0); i < tabCount; i++ {
		if i == m.activeTab {
			tabs[i] = tabActiveStyle.Render(i.String())
		} else {
			tabs[i] = tabInactiveStyle.Render(i.String())
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(tabs, "  "))
}

func (m stackDetailModel) ShortHelp() []key.Binding {
	bindings := []key.Binding{appKeys.Tab, appKeys.Back, appKeys.Refresh, appKeys.Quit}
	if m.activeTab == tabResources || m.activeTab == tabOperations {
		bindings = append([]key.Binding{appKeys.Select}, bindings...)
	}
	return bindings
}

func (m stackDetailModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

func (m stackDetailModel) updateFocus() stackDetailModel {
	m.resourceTable.Blur()
	m.operationTable.Blur()
	switch m.activeTab {
	case tabResources:
		m.resourceTable.Focus()
	case tabOperations:
		m.operationTable.Focus()
	}
	return m
}

func (m stackDetailModel) selectedResource() (api.Resource, bool) {
	idx := m.resourceTable.Cursor()
	if idx < 0 || idx >= len(m.resources) {
		return api.Resource{}, false
	}
	return m.resources[idx], true
}

func (m stackDetailModel) selectedOperation() (api.Operation, bool) {
	idx := m.operationTable.Cursor()
	if idx < 0 || idx >= len(m.operations) {
		return api.Operation{}, false
	}
	return m.operations[idx], true
}

func (m stackDetailModel) isLoading() bool {
	return m.loadingResources || m.loadingOperations || m.loadingLogs
}

func (m stackDetailModel) fetchResources() tea.Cmd {
	return func() tea.Msg {
		resources, err := m.client.ListResources(m.stack.ID)
		if err != nil {
			return apiErrMsg{err: err}
		}
		return resourcesLoadedMsg{resources: resources}
	}
}

func (m stackDetailModel) fetchOperations() tea.Cmd {
	return func() tea.Msg {
		ops, err := m.client.ListOperations(m.stack.ID, api.ListOperationsOpts{})
		if err != nil {
			return apiErrMsg{err: err}
		}
		return operationsLoadedMsg{operations: ops}
	}
}

func (m stackDetailModel) fetchLogs() tea.Cmd {
	return func() tea.Msg {
		logs, err := m.client.ListLogs(api.ListLogsOpts{StackID: m.stack.ID})
		if err != nil {
			return apiErrMsg{err: err}
		}
		return logsLoadedMsg{logs: logs}
	}
}

func formatLogs(logs []api.Log) string {
	if len(logs) == 0 {
		return mutedStyle.Render("No logs available.")
	}
	var b strings.Builder
	for i := len(logs) - 1; i >= 0; i-- {
		l := logs[i]
		ts := mutedStyle.Render(l.Timestamp.Format("2006-01-02 15:04:05"))
		level := logLevelStyle(l.Level).Render(fmt.Sprintf("%-5s", l.Level))
		b.WriteString(fmt.Sprintf("%s %s %s\n", ts, level, l.Message))
	}
	return b.String()
}
