package tui

import (
	"fmt"
	"sort"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sanity-labs/blueprints-tui/internal/api"
)

type scopeSelectedMsg struct {
	scopeType string
	scopeID   string
	label     string
}

type scopeDataMsg struct {
	orgs     []api.Organization
	projects []api.Project
}

type orgItem struct {
	org          api.Organization
	projectCount int
}

func (i orgItem) Title() string { return i.org.Name }
func (i orgItem) Description() string {
	return fmt.Sprintf("Organization  •  %s  •  %d projects", i.org.ID, i.projectCount)
}
func (i orgItem) FilterValue() string { return i.org.Name }

type projectItem struct {
	project api.Project
}

func (i projectItem) Title() string       { return "    " + i.project.DisplayName }
func (i projectItem) Description() string { return "    Project  •  " + i.project.ID }
func (i projectItem) FilterValue() string { return i.project.DisplayName }

type scopePickerModel struct {
	list    list.Model
	client  *api.Client
	styles  styles
	loading bool
	spinner spinner.Model
	err     error
	height  int
}

func newScopePickerModel(client *api.Client, s styles) scopePickerModel {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles.TitleBar = lipgloss.NewStyle()

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return scopePickerModel{
		list:    l,
		client:  client,
		styles:  s,
		loading: true,
		spinner: sp,
	}
}

func (m scopePickerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchScopeData())
}

func (m scopePickerModel) Update(msg tea.Msg) (scopePickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case scopeDataMsg:
		m.loading = false
		items := buildScopeItems(msg.orgs, msg.projects)
		cmd := m.list.SetItems(items)
		return m, cmd

	case apiErrMsg:
		m.loading = false
		m.err = msg.err

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	if !m.loading {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View returns exactly m.height lines. The list bubble renders at its
// SetSize height; loading/error states are placed in the same box.
func (m scopePickerModel) View() string {
	if m.err != nil {
		s := m.styles.title.Render("Error") + "\n\n" + m.err.Error()
		return lipgloss.PlaceVertical(m.height, lipgloss.Top, s)
	}
	if m.loading {
		s := m.spinner.View() + " Loading organizations…"
		return lipgloss.PlaceVertical(m.height, lipgloss.Top, s)
	}
	return m.list.View()
}

func (m *scopePickerModel) SetSize(w, h int) {
	m.height = h
	m.list.SetSize(w, h)
}

func (m scopePickerModel) selectedScope() (scopeSelectedMsg, bool) {
	item := m.list.SelectedItem()
	if item == nil {
		return scopeSelectedMsg{}, false
	}
	switch i := item.(type) {
	case orgItem:
		return scopeSelectedMsg{
			scopeType: "organization",
			scopeID:   i.org.ID,
			label:     i.org.Name,
		}, true
	case projectItem:
		return scopeSelectedMsg{
			scopeType: "project",
			scopeID:   i.project.ID,
			label:     i.project.DisplayName,
		}, true
	}
	return scopeSelectedMsg{}, false
}

func (m scopePickerModel) fetchScopeData() tea.Cmd {
	return func() tea.Msg {
		orgs, err := m.client.ListOrganizations()
		if err != nil {
			return apiErrMsg{err: err}
		}
		projects, err := m.client.ListProjects()
		if err != nil {
			return apiErrMsg{err: err}
		}
		return scopeDataMsg{orgs: orgs, projects: projects}
	}
}

func buildScopeItems(orgs []api.Organization, projects []api.Project) []list.Item {
	projectsByOrg := make(map[string][]api.Project)
	for _, p := range projects {
		projectsByOrg[p.OrganizationID] = append(projectsByOrg[p.OrganizationID], p)
	}

	sort.Slice(orgs, func(i, j int) bool {
		return orgs[i].Name < orgs[j].Name
	})

	var items []list.Item
	for _, org := range orgs {
		orgProjects := projectsByOrg[org.ID]
		items = append(items, orgItem{org: org, projectCount: len(orgProjects)})

		sort.Slice(orgProjects, func(i, j int) bool {
			return orgProjects[i].DisplayName < orgProjects[j].DisplayName
		})
		for _, p := range orgProjects {
			items = append(items, projectItem{project: p})
		}
	}

	return items
}
