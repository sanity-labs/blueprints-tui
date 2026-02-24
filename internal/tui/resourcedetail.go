package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sanity-labs/blueprints-tui/internal/api"
)

const resourceHeaderChrome = 2 // type line + blank line

type resourceDetailModel struct {
	resource api.Resource
	viewport viewport.Model
}

func newResourceDetailModel(r api.Resource, width, height int) resourceDetailModel {
	vp := viewport.New(width, height-resourceHeaderChrome)
	vp.SetContent(formatResource(r))

	return resourceDetailModel{
		resource: r,
		viewport: vp,
	}
}

func (m *resourceDetailModel) SetSize(w, h int) {
	m.viewport.Width = w
	m.viewport.Height = h - resourceHeaderChrome
}

func (m resourceDetailModel) Init() tea.Cmd {
	return nil
}

func (m resourceDetailModel) Update(msg tea.Msg) (resourceDetailModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m resourceDetailModel) View() string {
	var b strings.Builder

	b.WriteString(mutedStyle.Render(m.resource.Type) + "\n\n")
	b.WriteString(m.viewport.View())

	return b.String()
}

func formatResource(r api.Resource) string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("Parameters") + "\n")
	b.WriteString(formatMap(r.Parameters))
	b.WriteString("\n")

	if len(r.ProviderMetadata) > 0 {
		b.WriteString(headerStyle.Render("Provider Metadata") + "\n")
		b.WriteString(formatMap(r.ProviderMetadata))
	}

	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("Created: %s", r.CreatedAt.Format("2006-01-02 15:04:05"))))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("Updated: %s", r.UpdatedAt.Format("2006-01-02 15:04:05"))))

	if r.ExternalID != "" {
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render(fmt.Sprintf("External ID: %s", r.ExternalID)))
	}

	return b.String()
}

func formatMap(m map[string]any) string {
	if len(m) == 0 {
		return mutedStyle.Render("  (empty)") + "\n"
	}
	data, err := json.MarshalIndent(m, "  ", "  ")
	if err != nil {
		return fmt.Sprintf("  %v\n", m)
	}
	return "  " + string(data) + "\n"
}
