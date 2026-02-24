package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sanity-labs/blueprints-tui/internal/api"
)

type apiErrMsg struct {
	err error
}

type view int

const (
	viewScopePicker view = iota
	viewStackList
	viewStackDetail
	viewResourceDetail
	viewOperationDetail
)

type Model struct {
	client          *api.Client
	currentView     view
	scopePicker     scopePickerModel
	scopeLabel      string
	scopeType       string
	stackList       stackListModel
	stackDetail     stackDetailModel
	resourceDetail  resourceDetailModel
	operationDetail operationDetailModel
	help            help.Model
	showHelp        bool
	width           int
	height          int
}

func NewModel(client *api.Client, hasScope bool) Model {
	m := Model{
		client: client,
		help:   help.New(),
	}
	if hasScope {
		m.currentView = viewStackList
		m.stackList = newStackListModel(client)
	} else {
		m.currentView = viewScopePicker
		m.scopePicker = newScopePickerModel(client)
	}
	return m
}

func (m Model) Init() tea.Cmd {
	switch m.currentView {
	case viewScopePicker:
		return m.scopePicker.Init()
	default:
		return m.stackList.Init()
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.resizeCurrentView()
		return m, nil

	case scopeSelectedMsg:
		m.scopeLabel = msg.label
		m.scopeType = msg.scopeType
		m.client.SetScope(msg.scopeType, msg.scopeID)
		m.stackList = newStackListModel(m.client)
		m.stackList.SetSize(m.effectiveWidth(), m.contentHeight())
		m.currentView = viewStackList
		return m, m.stackList.Init()

	case tea.KeyMsg:
		if key.Matches(msg, appKeys.Quit) && !m.isFiltering() {
			return m, tea.Quit
		}
		if key.Matches(msg, appKeys.Help) {
			m.showHelp = !m.showHelp
			return m, nil
		}
		if nav, cmd, handled := m.handleNavigation(msg); handled {
			return nav, cmd
		}
	}

	return m.updateCurrentView(msg)
}

func (m Model) View() string {
	header := m.headerBox()

	var content string
	switch m.currentView {
	case viewScopePicker:
		content = m.scopePicker.View()
	case viewStackList:
		content = m.stackList.View()
	case viewStackDetail:
		content = m.stackDetail.View()
	case viewResourceDetail:
		content = m.resourceDetail.View()
	case viewOperationDetail:
		content = m.operationDetail.View()
	}

	top := lipgloss.JoinVertical(lipgloss.Left, header, "", content)

	if m.showHelp {
		top += "\n" + helpStyle.Render(m.help.View(appKeys))
	}

	status := m.statusBar()
	topH := lipgloss.Height(top)
	statusH := lipgloss.Height(status)
	gap := m.effectiveHeight() - topH - statusH
	if gap < 0 {
		gap = 0
	}

	return top + "\n" + strings.Repeat("\n", gap) + status
}

func (m Model) headerBox() string {
	title := headerTitleStyle.Render("Blueprints")
	sep := headerHintStyle.Render("  ▸  ")
	dot := headerHintStyle.Render("  ·  ")

	content := title

	switch m.currentView {
	case viewScopePicker:
		content += sep + headerHintStyle.Render("Select a scope")
	case viewStackList:
		content += sep + headerValueStyle.Render(m.scopeLabel)
	case viewStackDetail:
		content += sep + headerValueStyle.Render(m.scopeLabel) +
			dot + headerValueStyle.Render(m.stackDetail.stack.Name)
	case viewResourceDetail:
		content += sep + headerHintStyle.Render(m.scopeLabel) +
			dot + headerHintStyle.Render(m.stackDetail.stack.Name) +
			dot + headerValueStyle.Render(m.resourceDetail.resource.Name)
	case viewOperationDetail:
		content += sep + headerHintStyle.Render(m.scopeLabel) +
			dot + headerHintStyle.Render(m.stackDetail.stack.Name) +
			dot + headerValueStyle.Render(m.operationDetail.operation.ID)
	}

	return headerBoxStyle.Render(content)
}

func helpItem(key, label string) string {
	return keycapStyle.Render(key) + " " + headerHintStyle.Render(label)
}

func (m Model) statusBar() string {
	sep := headerHintStyle.Render("  ·  ")
	var hints []string
	switch m.currentView {
	case viewScopePicker:
		hints = []string{helpItem("ENTER", "select"), helpItem("/", "filter"), helpItem("?", "help"), helpItem("q", "quit")}
	case viewStackList:
		hints = []string{helpItem("ENTER", "select"), helpItem("/", "filter"), helpItem("r", "refresh"), helpItem("ESC", "back"), helpItem("?", "help"), helpItem("q", "quit")}
	case viewStackDetail:
		hints = []string{helpItem("ENTER", "select"), helpItem("TAB", "tabs"), helpItem("r", "refresh"), helpItem("ESC", "back"), helpItem("?", "help"), helpItem("q", "quit")}
	case viewResourceDetail, viewOperationDetail:
		hints = []string{helpItem("ESC", "back"), helpItem("?", "help"), helpItem("q", "quit")}
	}
	return strings.Join(hints, sep)
}

func (m Model) handleNavigation(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	switch m.currentView {
	case viewScopePicker:
		if m.scopePicker.list.FilterState() == list.Filtering {
			break
		}
		if key.Matches(msg, appKeys.Select) {
			if scope, ok := m.scopePicker.selectedScope(); ok {
				return m, func() tea.Msg { return scope }, true
			}
		}

	case viewStackList:
		if m.isFiltering() {
			break
		}
		if key.Matches(msg, appKeys.Back) && m.scopePicker.client != nil {
			m.currentView = viewScopePicker
			m.resizeCurrentView()
			return m, nil, true
		}
		if key.Matches(msg, appKeys.Select) {
			if stack, ok := m.stackList.selectedStack(); ok {
				m.currentView = viewStackDetail
				m.stackDetail = newStackDetailModel(m.client, stack, m.effectiveWidth(), m.contentHeight())
				return m, m.stackDetail.Init(), true
			}
		}
		if key.Matches(msg, appKeys.Refresh) {
			var cmd tea.Cmd
			m.stackList, cmd = m.stackList.Refresh()
			return m, cmd, true
		}

	case viewStackDetail:
		if key.Matches(msg, appKeys.Back) {
			m.currentView = viewStackList
			m.resizeCurrentView()
			return m, nil, true
		}
		if key.Matches(msg, appKeys.Select) {
			w, h := m.effectiveWidth(), m.contentHeight()
			switch m.stackDetail.activeTab {
			case tabResources:
				if r, ok := m.stackDetail.selectedResource(); ok {
					m.currentView = viewResourceDetail
					m.resourceDetail = newResourceDetailModel(r, w, h)
					return m, m.resourceDetail.Init(), true
				}
			case tabOperations:
				if op, ok := m.stackDetail.selectedOperation(); ok {
					m.currentView = viewOperationDetail
					m.operationDetail = newOperationDetailModel(m.client, m.stackDetail.stack.ID, op, w, h)
					return m, m.operationDetail.Init(), true
				}
			}
		}

	case viewResourceDetail:
		if key.Matches(msg, appKeys.Back) {
			m.currentView = viewStackDetail
			m.resizeCurrentView()
			return m, nil, true
		}

	case viewOperationDetail:
		if key.Matches(msg, appKeys.Back) {
			m.currentView = viewStackDetail
			m.resizeCurrentView()
			return m, nil, true
		}
	}

	return m, nil, false
}

func (m Model) updateCurrentView(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentView {
	case viewScopePicker:
		m.scopePicker, cmd = m.scopePicker.Update(msg)
	case viewStackList:
		m.stackList, cmd = m.stackList.Update(msg)
	case viewStackDetail:
		m.stackDetail, cmd = m.stackDetail.Update(msg)
	case viewResourceDetail:
		m.resourceDetail, cmd = m.resourceDetail.Update(msg)
	case viewOperationDetail:
		m.operationDetail, cmd = m.operationDetail.Update(msg)
	}
	return m, cmd
}

func (m *Model) resizeCurrentView() {
	w, h := m.effectiveWidth(), m.contentHeight()
	switch m.currentView {
	case viewScopePicker:
		m.scopePicker.SetSize(w, h)
	case viewStackList:
		m.stackList.SetSize(w, h)
	case viewStackDetail:
		m.stackDetail.SetSize(w, h)
	case viewResourceDetail:
		m.resourceDetail.SetSize(w, h)
	case viewOperationDetail:
		m.operationDetail.SetSize(w, h)
	}
}

func (m Model) isFiltering() bool {
	switch m.currentView {
	case viewScopePicker:
		return m.scopePicker.list.FilterState() == list.Filtering
	case viewStackList:
		return m.stackList.list.FilterState() == list.Filtering
	}
	return false
}

func (m Model) effectiveWidth() int {
	if m.width > 0 {
		return m.width
	}
	return 80
}

func (m Model) effectiveHeight() int {
	if m.height > 0 {
		return m.height
	}
	return 24
}

func (m Model) contentHeight() int {
	headerH := lipgloss.Height(m.headerBox())
	spacing := 1 // blank line between header and content
	statusH := 1
	h := m.effectiveHeight() - headerH - spacing - statusH
	if h < 1 {
		return 1
	}
	return h
}
