package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	listView   string = "list"
	detailView string = "detail"
)

type model struct {
	list list.Model

	lg *lipgloss.Renderer
	s  *Styles
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }
func (i item) Render() string {
	return i.title
}

func NewModel() *model {
	items := []list.Item{
		item{title: "Item 1", desc: "Description 1"},
		item{title: "Item 2", desc: "Description 2"},
		item{title: "Item 3", desc: "Description 3"},
	}

	m := &model{
		list: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}

	m.list.Title = "Items"

	m.lg = lipgloss.DefaultRenderer()
	m.s = NewStyles(m.lg)

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var body string

	body += m.renderHeader()

	body += "\n\n"
	body += m.list.View()

	return m.s.Base.Render(body)
}

func (m model) renderHeader() string {
	return m.s.HeaderText.Render("FWD Management")
}
