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
	viewStackList view = iota
	viewStackDetail
	viewResourceDetail
	viewOperationDetail
)

type Model struct {
	client          *api.Client
	currentView     view
	stackList       stackListModel
	stackDetail     stackDetailModel
	resourceDetail  resourceDetailModel
	operationDetail operationDetailModel
	help            help.Model
	showHelp        bool
	width           int
	height          int
}

func NewModel(client *api.Client) Model {
	return Model{
		client:      client,
		currentView: viewStackList,
		stackList:   newStackListModel(client),
		help:        help.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.stackList.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.SetWidth(msg.Width)

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
	crumbs := mutedStyle.Render("Stacks")

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
	case viewStackList:
		if m.isFiltering() {
			break
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
	return m.currentView == viewStackList && m.stackList.list.FilterState() == list.Filtering
}
