package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"github.com/sanity-io/blueprints-tui/internal/api"
)

const stackListChrome = 1 // status bar

type stackItem struct {
	stack api.Stack
}

func (i stackItem) Title() string       { return i.stack.Name }
func (i stackItem) Description() string { return fmt.Sprintf("%s  •  %s", i.stack.ID, i.stack.BlueprintID) }
func (i stackItem) FilterValue() string { return i.stack.Name }

type stacksLoadedMsg struct {
	stacks []api.Stack
}

type apiErrMsg struct {
	err error
}

type stackListModel struct {
	list    list.Model
	client  *api.Client
	loading bool
	spinner spinner.Model
	err     error
	width   int
	height  int
}

func newStackListModel(client *api.Client) stackListModel {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.Title = "Blueprints Stacks"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-stackListChrome)

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


func (m *stackListModel) Refresh() tea.Cmd {
	m.loading = true
	return tea.Batch(m.spinner.Tick, m.fetchStacks())
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
