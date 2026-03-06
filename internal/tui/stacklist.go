package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sanity-labs/blueprints-tui/internal/api"
)

type stackItem struct {
	stack  api.Stack
	styles *styles
}

func (i stackItem) Title() string { return i.stack.Name }
func (i stackItem) Description() string {
	var parts []string
	parts = append(parts, i.stack.ID)
	if count := i.stack.DisplayResourceCount(); count != nil {
		n := *count
		if n == 1 {
			parts = append(parts, "1 resource")
		} else {
			parts = append(parts, fmt.Sprintf("%d resources", n))
		}
	}
	parts = append(parts, i.stack.BlueprintID)
	desc := strings.Join(parts, "  ·  ")
	if op := i.stack.RecentOperation; op != nil && i.styles != nil {
		desc = i.styles.statusIndicator(op.Status) + " " + desc
	}
	return desc
}
func (i stackItem) FilterValue() string {
	return i.stack.Name + " " + i.stack.ID + " " + i.stack.BlueprintID
}

type stacksLoadedMsg struct {
	stacks []api.Stack
}

type stackListModel struct {
	list    list.Model
	client  *api.Client
	styles  styles
	stacks  []api.Stack
	loading bool
	spinner spinner.Model
	err     error
	height  int
}

func newStackListModel(client *api.Client, s styles) stackListModel {
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

	return stackListModel{
		list:    l,
		client:  client,
		styles:  s,
		loading: true,
		spinner: sp,
	}
}

func (m stackListModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchStacks())
}

func (m stackListModel) Update(msg tea.Msg) (stackListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case stacksLoadedMsg:
		m.loading = false
		m.err = nil
		m.stacks = msg.stacks
		items := make([]list.Item, len(msg.stacks))
		for i, s := range msg.stacks {
			items[i] = stackItem{stack: s, styles: &m.styles}
		}
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
func (m stackListModel) View() string {
	if m.err != nil {
		s := m.styles.title.Render("Error") + "\n\n" + m.err.Error() + "\n\n" + m.styles.muted.Render("Press r to retry")
		return lipgloss.PlaceVertical(m.height, lipgloss.Top, s)
	}
	if m.loading {
		s := m.spinner.View() + " Loading stacks…"
		return lipgloss.PlaceVertical(m.height, lipgloss.Top, s)
	}
	if len(m.list.Items()) == 0 {
		s := m.styles.muted.Render("No stacks found.")
		return lipgloss.PlaceVertical(m.height, lipgloss.Top, s)
	}
	return m.list.View()
}

func (m *stackListModel) SetSize(w, h int) {
	m.height = h
	m.list.SetSize(w, h)
}

func (m stackListModel) selectedStack() (api.Stack, bool) {
	item := m.list.SelectedItem()
	if item == nil {
		return api.Stack{}, false
	}
	si, ok := item.(stackItem)
	if !ok {
		return api.Stack{}, false
	}
	return si.stack, true
}

func (m stackListModel) Refresh() (stackListModel, tea.Cmd) {
	m.loading = true
	m.err = nil
	return m, tea.Batch(m.spinner.Tick, m.fetchStacks())
}

func (m stackListModel) fetchStacks() tea.Cmd {
	return func() tea.Msg {
		stacks, err := m.client.ListStacks()
		if err != nil {
			return apiErrMsg{err: err}
		}
		return stacksLoadedMsg{stacks: stacks}
	}
}
