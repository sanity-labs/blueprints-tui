package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sanity-labs/blueprints-tui/internal/api"
)

const resourceHeaderChrome = 3 // name + type + blank line

type resourceLoadedMsg struct {
	resource api.Resource
}

type resourceDetailModel struct {
	resource     api.Resource
	fullResource *api.Resource
	client       *api.Client
	stackID      string
	styles       styles
	viewport     viewport.Model
	spinner      spinner.Model
	loading      bool
	err          error
	height       int
}

func newResourceDetailModel(client *api.Client, stackID string, r api.Resource, s styles, width, height int) resourceDetailModel {
	vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height-resourceHeaderChrome))
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return resourceDetailModel{
		resource: r,
		client:   client,
		stackID:  stackID,
		styles:   s,
		viewport: vp,
		spinner:  sp,
		loading:  true,
		height:   height,
	}
}

func (m *resourceDetailModel) SetSize(w, h int) {
	m.height = h
	m.viewport.SetWidth(w)
	m.viewport.SetHeight(h - resourceHeaderChrome)
}

func (m resourceDetailModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchResource())
}

func (m resourceDetailModel) Update(msg tea.Msg) (resourceDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case resourceLoadedMsg:
		m.loading = false
		m.fullResource = &msg.resource
		m.viewport.SetContent(m.formatResource(msg.resource))
		return m, nil

	case apiErrMsg:
		m.loading = false
		m.err = msg.err
		m.viewport.SetContent(m.formatResource(m.resource))
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View returns exactly m.height lines. Chrome (name + type + blank) is
// fixed; the inner area (viewport or spinner) is placed to fill the rest.
func (m resourceDetailModel) View() string {
	s := m.styles
	r := m.displayResource()
	chrome := s.headerValue.Render(r.Name) + "\n" + s.muted.Render(r.Type) + "\n"

	var inner string
	if m.loading {
		inner = m.spinner.View() + " Loading resource…"
	} else {
		inner = m.viewport.View()
	}

	innerH := m.height - resourceHeaderChrome
	if innerH < 1 {
		innerH = 1
	}
	inner = lipgloss.PlaceVertical(innerH, lipgloss.Top, inner)

	return chrome + "\n" + inner
}

func (m resourceDetailModel) displayResource() api.Resource {
	if m.fullResource != nil {
		return *m.fullResource
	}
	return m.resource
}

func (m resourceDetailModel) fetchResource() tea.Cmd {
	return func() tea.Msg {
		r, err := m.client.GetResource(m.stackID, m.resource.ID)
		if err != nil {
			return apiErrMsg{err: err}
		}
		return resourceLoadedMsg{resource: r}
	}
}

func (m resourceDetailModel) formatResource(r api.Resource) string {
	s := m.styles
	var b strings.Builder

	b.WriteString(s.sectionHead.Render("Parameters") + "\n")
	b.WriteString(formatMap(r.Parameters))
	b.WriteString("\n")

	if len(r.ProviderMetadata) > 0 {
		b.WriteString(s.sectionHead.Render("Provider Metadata") + "\n")
		b.WriteString(formatMap(r.ProviderMetadata))
		b.WriteString("\n")
	}

	b.WriteString(s.muted.Render(strings.Repeat("─", 40)) + "\n\n")

	writeRow := func(label, value string) {
		b.WriteString(fmt.Sprintf("  %s  %s\n", s.muted.Render(fmt.Sprintf("%-14s", label)), value))
	}

	writeRow("Created", r.CreatedAt.Format("2006-01-02 15:04"))
	writeRow("Updated", r.UpdatedAt.Format("2006-01-02 15:04"))
	if r.ExternalID != "" {
		writeRow("External ID", r.ExternalID)
	}
	writeRow("Resource ID", r.ID)
	writeRow("Operation ID", r.OperationID)

	return b.String()
}

func formatMap(m map[string]any) string {
	if len(m) == 0 {
		return "  (empty)\n"
	}
	data, err := json.MarshalIndent(m, "  ", "  ")
	if err != nil {
		return fmt.Sprintf("  %v\n", m)
	}
	return "  " + string(data) + "\n"
}
