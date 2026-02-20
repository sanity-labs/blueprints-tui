package tui

import (
	"fmt"
	"sort"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"github.com/sanity-io/blueprints-tui/internal/api"
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

func (i orgItem) Title() string       { return i.org.Name }
func (i orgItem) Description() string { return fmt.Sprintf("Organization  •  %s  •  %d projects", i.org.ID, i.projectCount) }
func (i orgItem) FilterValue() string { return i.org.Name }

type projectItem struct {
	project api.Project
}

func (i projectItem) Title() string       { return "  ↳ " + i.project.DisplayName }
func (i projectItem) Description() string { return "    Project  •  " + i.project.ID }
func (i projectItem) FilterValue() string { return i.project.DisplayName }

const scopePickerChrome = 2

type scopePickerModel struct {
	list    list.Model
	client  *api.Client
	loading bool
	spinner spinner.Model
	err     error
	width   int
	height  int
}

func newScopePickerModel(client *api.Client) scopePickerModel {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.Title = "Select Scope"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	sp := spinner.New(spinner.WithSpinner(spinner.Dot))

	return scopePickerModel{
		list:    l,
		client:  client,
		loading: true,
		spinner: sp,
	}
}

func (m scopePickerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchScopeData())
}

func (m scopePickerModel) Update(msg tea.Msg) (scopePickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-scopePickerChrome)

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

func (m scopePickerModel) View() string {
	if m.err != nil {
		return titleStyle.Render("Error") + "\n\n" + m.err.Error()
	}
	if m.loading {
		return m.spinner.View() + " Loading organizations…"
	}

	out := m.list.View()

	if item := m.list.SelectedItem(); item != nil {
		if _, ok := item.(projectItem); ok {
			out += "\n" + mutedStyle.Render("  Project scope is being phased out. Consider selecting the organization instead.")
		}
	}

	return out
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
