package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sanity-labs/blueprints-tui/internal/api"
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

type stackLoadedMsg struct {
	stack api.Stack
}

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
	fullStack  *api.Stack
	client     *api.Client
	styles     styles
	activeTab  detailTab
	resources  []api.Resource
	operations []api.Operation
	logs       []api.Log

	resourceTable  table.Model
	operationTable table.Model
	logViewport    viewport.Model
	spinner        spinner.Model

	loadingStack      bool
	loadingResources  bool
	loadingOperations bool
	loadingLogs       bool
	resourcesLoaded   bool
	operationsLoaded  bool
	logsLoaded        bool
	err               error
	width             int
	height            int
}

// chromeHeight returns the measured height of the non-scrollable region
// (stack header + tab bar). This is always stable regardless of data.
func (m stackDetailModel) chromeHeight() int {
	return lipgloss.Height(m.renderHeader()) + lipgloss.Height(m.renderTabs())
}

// innerHeight returns the space available for the tab's scrollable content.
func (m stackDetailModel) innerHeight() int {
	h := m.height - m.chromeHeight()
	if h < 1 {
		return 1
	}
	return h
}

func newStackDetailModel(client *api.Client, stack api.Stack, s styles, width, height int) stackDetailModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := stackDetailModel{
		stack:            stack,
		client:           client,
		styles:           s,
		activeTab:        tabResources,
		spinner:          sp,
		width:            width,
		height:           height,
		loadingStack:     true,
		loadingResources: true,
	}

	innerH := m.innerHeight()

	rt := table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 30},
			{Title: "Type", Width: 30},
			{Title: "ID", Width: 16},
		}),
		table.WithFocused(true),
		table.WithWidth(width),
		table.WithHeight(innerH),
	)

	ot := table.New(
		table.WithColumns([]table.Column{
			{Title: " ", Width: 3},
			{Title: "ID", Width: 16},
			{Title: "Status", Width: 14},
			{Title: "Created", Width: 20},
		}),
		table.WithWidth(width),
		table.WithHeight(innerH),
	)

	rt.SetStyles(s.table)
	ot.SetStyles(s.table)

	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(innerH))

	m.resourceTable = rt
	m.operationTable = ot
	m.logViewport = vp
	return m
}

func (m stackDetailModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchStack(),
		m.fetchResources(),
	)
}

func (m stackDetailModel) Update(msg tea.Msg) (stackDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, appKeys.Tab):
			m.activeTab = (m.activeTab + 1) % tabCount
			m = m.updateFocus()
			return m, m.ensureTabLoaded()
		case key.Matches(msg, appKeys.ShiftTab):
			m.activeTab = (m.activeTab + tabCount - 1) % tabCount
			m = m.updateFocus()
			return m, m.ensureTabLoaded()
		case key.Matches(msg, appKeys.Refresh):
			return m, m.refreshTab()
		}

	case stackLoadedMsg:
		m.loadingStack = false
		m.fullStack = &msg.stack

	case resourcesLoadedMsg:
		m.loadingResources = false
		m.resourcesLoaded = true
		m.resources = msg.resources
		rows := make([]table.Row, len(msg.resources))
		for i, r := range msg.resources {
			rows[i] = table.Row{r.Name, r.Type, r.ID}
		}
		m.resourceTable.SetRows(rows)

	case operationsLoadedMsg:
		m.loadingOperations = false
		m.operationsLoaded = true
		m.operations = msg.operations
		rows := make([]table.Row, len(msg.operations))
		for i, op := range msg.operations {
			rows[i] = table.Row{
				m.styles.statusIndicator(op.Status),
				op.ID,
				op.Status,
				op.CreatedAt.Format("2006-01-02 15:04"),
			}
		}
		m.operationTable.SetRows(rows)

	case logsLoadedMsg:
		m.loadingLogs = false
		m.logsLoaded = true
		m.logs = msg.logs
		m.logViewport.SetContent(m.formatLogs(msg.logs))
		m.logViewport.GotoBottom()

	case apiErrMsg:
		m.err = msg.err
		m.loadingStack = false
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

// View returns exactly m.height lines. Chrome (header + tabs) is fixed;
// the inner content area is placed in a box of innerHeight lines so that
// loading spinners and empty states don't shift the layout.
func (m stackDetailModel) View() string {
	s := m.styles
	chrome := m.renderHeader() + m.renderTabs()

	var inner string
	if m.err != nil {
		inner = "Error: " + m.err.Error()
	} else {
		switch m.activeTab {
		case tabResources:
			if m.loadingResources {
				inner = m.spinner.View() + " Loading resources…"
			} else if len(m.resources) == 0 {
				inner = s.muted.Render("No resources.")
			} else {
				inner = m.resourceTable.View()
			}
		case tabOperations:
			if m.loadingOperations {
				inner = m.spinner.View() + " Loading operations…"
			} else if len(m.operations) == 0 {
				inner = s.muted.Render("No operations.")
			} else {
				inner = m.operationTable.View()
			}
		case tabLogs:
			if m.loadingLogs {
				inner = m.spinner.View() + " Loading logs…"
			} else if len(m.logs) == 0 {
				inner = s.muted.Render("No logs available.")
			} else {
				inner = m.logViewport.View()
			}
		}
	}

	// Force inner to exactly innerHeight so the total is always m.height.
	inner = lipgloss.PlaceVertical(m.innerHeight(), lipgloss.Top, inner)

	return chrome + "\n" + inner
}

func (m stackDetailModel) renderHeader() string {
	s := m.styles
	ds := m.displayStack()

	line1 := s.muted.Render("Stack: ") + s.headerValue.Render(ds.Name)

	var meta []string
	meta = append(meta, ds.ID)
	if count := ds.DisplayResourceCount(); count != nil {
		n := *count
		if n == 1 {
			meta = append(meta, "1 resource")
		} else {
			meta = append(meta, fmt.Sprintf("%d resources", n))
		}
	}
	if op := ds.RecentOperation; op != nil {
		meta = append(meta, s.statusIndicator(op.Status)+" "+op.Status)
		meta = append(meta, op.CreatedAt.Format("2006-01-02 15:04"))
	}
	meta = append(meta, ds.BlueprintID)
	line2 := s.muted.Render(strings.Join(meta, "  ·  "))

	return line1 + "\n" + line2 + "\n"
}

func (m stackDetailModel) displayStack() api.Stack {
	if m.fullStack != nil {
		return *m.fullStack
	}
	return m.stack
}

func (m stackDetailModel) renderTabs() string {
	s := m.styles
	tabs := make([]string, tabCount)
	for i := detailTab(0); i < tabCount; i++ {
		if i == m.activeTab {
			tabs[i] = s.tabActive.Render(i.String())
		} else {
			tabs[i] = s.tabInactive.Render(i.String())
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(tabs, " "))
	w := m.width
	if w < 1 {
		w = lipgloss.Width(row)
	}
	underline := s.tabUnderline.Render(strings.Repeat("━", w))
	return row + "\n" + underline
}

func (m *stackDetailModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	innerH := m.innerHeight()
	m.resourceTable.SetWidth(w)
	m.resourceTable.SetHeight(innerH)
	m.operationTable.SetWidth(w)
	m.operationTable.SetHeight(innerH)
	m.logViewport.SetWidth(w)
	m.logViewport.SetHeight(innerH)
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

// ensureTabLoaded starts a fetch for the active tab if it hasn't loaded yet.
// Pointer receiver so the loading flag is visible to the caller's m.
func (m *stackDetailModel) ensureTabLoaded() tea.Cmd {
	switch m.activeTab {
	case tabResources:
		if !m.resourcesLoaded {
			m.loadingResources = true
			return tea.Batch(m.spinner.Tick, m.fetchResources())
		}
	case tabOperations:
		if !m.operationsLoaded {
			m.loadingOperations = true
			return tea.Batch(m.spinner.Tick, m.fetchOperations())
		}
	case tabLogs:
		if !m.logsLoaded {
			m.loadingLogs = true
			return tea.Batch(m.spinner.Tick, m.fetchLogs())
		}
	}
	return nil
}

func (m *stackDetailModel) refreshTab() tea.Cmd {
	switch m.activeTab {
	case tabResources:
		m.loadingResources = true
		return tea.Batch(m.spinner.Tick, m.fetchStack(), m.fetchResources())
	case tabOperations:
		m.loadingOperations = true
		return tea.Batch(m.spinner.Tick, m.fetchStack(), m.fetchOperations())
	case tabLogs:
		m.loadingLogs = true
		return tea.Batch(m.spinner.Tick, m.fetchStack(), m.fetchLogs())
	}
	return nil
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
	return m.loadingStack || m.loadingResources || m.loadingOperations || m.loadingLogs
}

func (m stackDetailModel) fetchStack() tea.Cmd {
	return func() tea.Msg {
		stack, err := m.client.GetStack(m.stack.ID)
		if err != nil {
			return apiErrMsg{err: err}
		}
		return stackLoadedMsg{stack: stack}
	}
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

func (m stackDetailModel) formatLogs(logs []api.Log) string {
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
