package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sanity-labs/blueprints-tui/internal/api"
)

type stackItem struct {
	stack api.Stack
}

func (i stackItem) Title() string { return i.stack.Name }
func (i stackItem) Description() string {
	return fmt.Sprintf("%s  •  %s", i.stack.ID, i.stack.BlueprintID)
}
func (i stackItem) FilterValue() string { return i.stack.Name }

type stacksLoadedMsg struct {
	stacks []api.Stack
}

type stackListModel struct {
	list    list.Model
	client  *api.Client
	loading bool
	spinner spinner.Model
	err     error
}

func newStackListModel(client *api.Client) stackListModel {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles.TitleBar = lipgloss.NewStyle()

	sp := spinner.New(spinner.WithSpinner(spinner.Dot))

	return stackListModel{
		list:    l,
		client:  client,
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
		items := make([]list.Item, len(msg.stacks))
		for i, s := range msg.stacks {
			items[i] = stackItem{stack: s}
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

func (m stackListModel) View() string {
	if m.err != nil {
		return titleStyle.Render("Error") + "\n\n" + m.err.Error()
	}
	if m.loading {
		return m.spinner.View() + " Loading stacks…"
	}
	return m.list.View()
}

func (m *stackListModel) SetSize(w, h int) {
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
