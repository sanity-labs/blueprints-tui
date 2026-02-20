package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"github.com/sanity-io/blueprints-tui/internal/api"
)

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
		m.help.SetWidth(msg.Width)

	case scopeSelectedMsg:
		m.scopeLabel = msg.label
		m.scopeType = msg.scopeType
		m.client.SetScope(msg.scopeType, msg.scopeID)
		m.stackList = newStackListModel(m.client)
		m.stackList.list.Title = "Stacks — " + msg.label
		m.stackList.width = m.width
		m.stackList.height = m.height
		m.stackList.list.SetSize(m.width, m.height-stackListChrome)
		m.currentView = viewStackList
		return m, m.stackList.Init()

	case tea.KeyPressMsg:
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

func (m Model) View() tea.View {
	var content string
	switch m.currentView {
	case viewScopePicker:
		content = m.scopePicker.View()
	case viewStackList:
		content = m.stackList.View()
	case viewStackDetail:
		content = m.breadcrumb() + m.stackDetail.View()
	case viewResourceDetail:
		content = m.breadcrumb() + m.resourceDetail.View()
	case viewOperationDetail:
		content = m.breadcrumb() + m.operationDetail.View()
	}

	if m.showHelp {
		content += "\n" + helpStyle.Render(m.help.View(appKeys))
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m Model) breadcrumb() string {
	sep := mutedStyle.Render(" › ")
	crumbs := titleStyle.Render(m.scopeLabel) + sep + mutedStyle.Render("Stacks")

	switch m.currentView {
	case viewStackDetail:
		crumbs += sep + titleStyle.Render(m.stackDetail.stack.Name) +
			" " + mutedStyle.Render(m.stackDetail.stack.ID)
	case viewResourceDetail:
		crumbs += sep + mutedStyle.Render(m.stackDetail.stack.Name) +
			sep + titleStyle.Render(m.resourceDetail.resource.Name) +
			" " + mutedStyle.Render(m.resourceDetail.resource.ID)
	case viewOperationDetail:
		crumbs += sep + mutedStyle.Render(m.stackDetail.stack.Name) +
			sep + titleStyle.Render(m.operationDetail.operation.ID) +
			" " + statusStyle(m.operationDetail.operation.Status).Render(m.operationDetail.operation.Status)
	}

	return crumbs + "\n"
}

func (m Model) handleNavigation(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
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
			m.scopePicker.list.SetSize(m.width, m.height-scopePickerChrome)
			m.currentView = viewScopePicker
			return m, nil, true
		}
		if key.Matches(msg, appKeys.Select) {
			if stack, ok := m.stackList.selectedStack(); ok {
				m.currentView = viewStackDetail
				m.stackDetail = newStackDetailModel(m.client, stack, m.width, m.height)
				return m, m.stackDetail.Init(), true
			}
		}
		if key.Matches(msg, appKeys.Refresh) {
			return m, m.stackList.Refresh(), true
		}

	case viewStackDetail:
		if key.Matches(msg, appKeys.Back) {
			m.currentView = viewStackList
			return m, nil, true
		}
		if key.Matches(msg, appKeys.Select) {
			switch m.stackDetail.activeTab {
			case tabResources:
				if r, ok := m.stackDetail.selectedResource(); ok {
					m.currentView = viewResourceDetail
					m.resourceDetail = newResourceDetailModel(r, m.width, m.height)
					return m, m.resourceDetail.Init(), true
				}
			case tabOperations:
				if op, ok := m.stackDetail.selectedOperation(); ok {
					m.currentView = viewOperationDetail
					m.operationDetail = newOperationDetailModel(m.client, m.stackDetail.stack.ID, op, m.width, m.height)
					return m, m.operationDetail.Init(), true
				}
			}
		}

	case viewResourceDetail:
		if key.Matches(msg, appKeys.Back) {
			m.currentView = viewStackDetail
			return m, nil, true
		}

	case viewOperationDetail:
		if key.Matches(msg, appKeys.Back) {
			m.currentView = viewStackDetail
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

func (m Model) isFiltering() bool {
	switch m.currentView {
	case viewScopePicker:
		return m.scopePicker.list.FilterState() == list.Filtering
	case viewStackList:
		return m.stackList.list.FilterState() == list.Filtering
	}
	return false
}
