package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sanity-labs/blueprints-tui/internal/api"
	"strings"
)

type apiErrMsg struct {
	err error
}

type route int

const (
	routeScopePicker route = iota
	routeStackList
	routeStackDetail
	routeResourceDetail
	routeOperationDetail
)

type Model struct {
	client *api.Client
	styles styles
	nav    []route

	scopePicker     scopePickerModel
	scopeLabel      string
	scopeType       string
	stackList       stackListModel
	stackDetail     stackDetailModel
	resourceDetail  resourceDetailModel
	operationDetail operationDetailModel

	help     help.Model
	showHelp bool
	width    int
	height   int
}

func NewModel(client *api.Client, hasScope bool) Model {
	s := newStyles(true)
	m := Model{
		client: client,
		styles: s,
		help:   help.New(),
	}
	if hasScope {
		m.nav = []route{routeStackList}
		m.stackList = newStackListModel(client, s)
	} else {
		m.nav = []route{routeScopePicker}
		m.scopePicker = newScopePickerModel(client, s)
	}
	return m
}

func (m Model) currentRoute() route {
	if len(m.nav) == 0 {
		return routeScopePicker
	}
	return m.nav[len(m.nav)-1]
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.RequestBackgroundColor}
	switch m.currentRoute() {
	case routeScopePicker:
		cmds = append(cmds, m.scopePicker.Init())
	default:
		cmds = append(cmds, m.stackList.Init())
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.styles = newStyles(msg.IsDark())
		m.updateChildStyles()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.SetWidth(msg.Width)
		m.resizeCurrentView()
		return m, nil

	case scopeSelectedMsg:
		m.scopeLabel = msg.label
		m.scopeType = msg.scopeType
		m.client.SetScope(msg.scopeType, msg.scopeID)
		m.stackList = newStackListModel(m.client, m.styles)
		m.stackList.SetSize(m.effectiveWidth(), m.contentHeight())
		m.nav = append(m.nav, routeStackList)
		return m, m.stackList.Init()

	case tea.KeyPressMsg:
		if key.Matches(msg, appKeys.Quit) && !m.isFiltering() {
			return m, tea.Quit
		}
		if key.Matches(msg, appKeys.Help) {
			m.showHelp = !m.showHelp
			m.resizeCurrentView()
			return m, nil
		}
		if nav, cmd, handled := m.handleNavigation(msg); handled {
			return nav, cmd
		}
	}

	return m.updateCurrentView(msg)
}

// View assembles three fixed-size vertical regions:
//
//	header (\n\n) content (\n) footer
//
// Each sub-view guarantees it renders at exactly contentHeight lines.
// The footer is the status bar, optionally preceded by expanded help.
// No gap math, no PlaceVertical at this level.
func (m Model) View() tea.View {
	header := m.headerBox()
	footer := m.footerView()

	var content string
	switch m.currentRoute() {
	case routeScopePicker:
		content = m.scopePicker.View()
	case routeStackList:
		content = m.stackList.View()
	case routeStackDetail:
		content = m.stackDetail.View()
	case routeResourceDetail:
		content = m.resourceDetail.View()
	case routeOperationDetail:
		content = m.operationDetail.View()
	}

	v := tea.NewView(header + "\n\n" + content + "\n" + footer)
	v.AltScreen = true
	return v
}

// footerView returns the status bar, optionally preceded by expanded help.
func (m Model) footerView() string {
	status := m.statusBar()
	if m.showHelp {
		return m.styles.help.Render(m.help.View(appKeys)) + "\n" + status
	}
	return status
}

func (m Model) headerBox() string {
	s := m.styles
	title := s.headerTitle.Render("Blueprints")
	sep := s.headerHint.Render("  ▸  ")
	dot := s.headerHint.Render("  ·  ")

	c := title

	switch m.currentRoute() {
	case routeScopePicker:
		c += sep + s.headerHint.Render("Select a scope")
	case routeStackList:
		c += sep + s.headerValue.Render(m.scopeLabel)
	case routeStackDetail:
		c += sep + s.headerValue.Render(m.scopeLabel) +
			dot + s.headerValue.Render(m.stackDetail.stack.Name)
	case routeResourceDetail:
		c += sep + s.headerHint.Render(m.scopeLabel) +
			dot + s.headerHint.Render(m.stackDetail.stack.Name) +
			dot + s.headerValue.Render(m.resourceDetail.resource.Name)
	case routeOperationDetail:
		c += sep + s.headerHint.Render(m.scopeLabel) +
			dot + s.headerHint.Render(m.stackDetail.stack.Name) +
			dot + s.headerValue.Render(m.operationDetail.operation.ID)
	}

	return s.headerBox.Render(c)
}

func (m Model) helpItem(k, label string) string {
	return m.styles.keycap.Render(k) + " " + m.styles.headerHint.Render(label)
}

func (m Model) statusBar() string {
	sep := m.styles.headerHint.Render("  ·  ")
	var hints []string
	switch m.currentRoute() {
	case routeScopePicker:
		hints = []string{m.helpItem("ENTER", "select"), m.helpItem("/", "filter"), m.helpItem("?", "help"), m.helpItem("q", "quit")}
	case routeStackList:
		hints = []string{m.helpItem("ENTER", "select"), m.helpItem("/", "filter"), m.helpItem("r", "refresh"), m.helpItem("ESC", "back"), m.helpItem("?", "help"), m.helpItem("q", "quit")}
	case routeStackDetail:
		hints = []string{m.helpItem("ENTER", "select"), m.helpItem("TAB", "tabs"), m.helpItem("r", "refresh"), m.helpItem("ESC", "back"), m.helpItem("?", "help"), m.helpItem("q", "quit")}
	case routeResourceDetail, routeOperationDetail:
		hints = []string{m.helpItem("ESC", "back"), m.helpItem("?", "help"), m.helpItem("q", "quit")}
	}
	return strings.Join(hints, sep)
}

func (m Model) handleNavigation(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	cur := m.currentRoute()

	switch cur {
	case routeScopePicker:
		if m.scopePicker.list.FilterState() == list.Filtering {
			break
		}
		if key.Matches(msg, appKeys.Select) {
			if scope, ok := m.scopePicker.selectedScope(); ok {
				return m, func() tea.Msg { return scope }, true
			}
		}

	case routeStackList:
		if m.isFiltering() {
			break
		}
		if key.Matches(msg, appKeys.Back) && m.scopePicker.client != nil {
			m.nav = m.nav[:len(m.nav)-1]
			m.resizeCurrentView()
			return m, nil, true
		}
		if key.Matches(msg, appKeys.Select) {
			if stack, ok := m.stackList.selectedStack(); ok {
				m.stackDetail = newStackDetailModel(m.client, stack, m.styles, m.effectiveWidth(), m.contentHeight())
				m.nav = append(m.nav, routeStackDetail)
				return m, m.stackDetail.Init(), true
			}
		}
		if key.Matches(msg, appKeys.Refresh) {
			var cmd tea.Cmd
			m.stackList, cmd = m.stackList.Refresh()
			return m, cmd, true
		}

	case routeStackDetail:
		if key.Matches(msg, appKeys.Back) {
			m.nav = m.nav[:len(m.nav)-1]
			m.resizeCurrentView()
			return m, nil, true
		}
		if key.Matches(msg, appKeys.Select) {
			w, h := m.effectiveWidth(), m.contentHeight()
			switch m.stackDetail.activeTab {
			case tabResources:
				if r, ok := m.stackDetail.selectedResource(); ok {
					m.resourceDetail = newResourceDetailModel(m.client, m.stackDetail.stack.ID, r, m.styles, w, h)
					m.nav = append(m.nav, routeResourceDetail)
					return m, m.resourceDetail.Init(), true
				}
			case tabOperations:
				if op, ok := m.stackDetail.selectedOperation(); ok {
					m.operationDetail = newOperationDetailModel(m.client, m.stackDetail.stack.ID, op, m.styles, w, h)
					m.nav = append(m.nav, routeOperationDetail)
					return m, m.operationDetail.Init(), true
				}
			}
		}

	case routeResourceDetail:
		if key.Matches(msg, appKeys.Back) {
			m.nav = m.nav[:len(m.nav)-1]
			m.resizeCurrentView()
			return m, nil, true
		}

	case routeOperationDetail:
		if key.Matches(msg, appKeys.Back) {
			m.nav = m.nav[:len(m.nav)-1]
			m.resizeCurrentView()
			return m, nil, true
		}
	}

	return m, nil, false
}

func (m Model) updateCurrentView(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentRoute() {
	case routeScopePicker:
		m.scopePicker, cmd = m.scopePicker.Update(msg)
	case routeStackList:
		m.stackList, cmd = m.stackList.Update(msg)
	case routeStackDetail:
		m.stackDetail, cmd = m.stackDetail.Update(msg)
	case routeResourceDetail:
		m.resourceDetail, cmd = m.resourceDetail.Update(msg)
	case routeOperationDetail:
		m.operationDetail, cmd = m.operationDetail.Update(msg)
	}
	return m, cmd
}

func (m *Model) updateChildStyles() {
	s := m.styles
	m.scopePicker.styles = s
	m.stackList.styles = s
	m.stackDetail.styles = s
	m.resourceDetail.styles = s
	m.operationDetail.styles = s
}

func (m *Model) resizeCurrentView() {
	w, h := m.effectiveWidth(), m.contentHeight()
	switch m.currentRoute() {
	case routeScopePicker:
		m.scopePicker.SetSize(w, h)
	case routeStackList:
		m.stackList.SetSize(w, h)
	case routeStackDetail:
		m.stackDetail.SetSize(w, h)
	case routeResourceDetail:
		m.resourceDetail.SetSize(w, h)
	case routeOperationDetail:
		m.operationDetail.SetSize(w, h)
	}
}

func (m Model) isFiltering() bool {
	switch m.currentRoute() {
	case routeScopePicker:
		return m.scopePicker.list.FilterState() == list.Filtering
	case routeStackList:
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

// contentHeight computes the vertical space available for the active sub-view.
// Layout: header \n\n content \n footer.
// Total = headerH + 1(blank) + contentH + footerH = effectiveHeight.
func (m Model) contentHeight() int {
	headerH := lipgloss.Height(m.headerBox())
	footerH := lipgloss.Height(m.footerView())
	h := m.effectiveHeight() - headerH - 1 - footerH
	if h < 1 {
		return 1
	}
	return h
}
