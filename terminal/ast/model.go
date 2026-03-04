package ast

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/midbel/dockit/formula/repr"
)

type NodeItem struct {
	node *repr.Node
}

func (i NodeItem) String() string {
	return i.node.Type
}

func (i NodeItem) Description() string {
	return i.node.Type
}

func (i NodeItem) Title() string {
	return fmt.Sprintf("▶ %s - %s", i.node.Name, i.node.Type)
}

func (i NodeItem) FilterValue() string {
	return ""
}

type astMsg struct {
	envelop *repr.Envelop
	err     error
}

type AstModel struct {
	width  int
	height int
	script string

	itemsList list.Model

	envelop *repr.Envelop
	err     error
}

func NewModel(script string) *AstModel {
	mod := &AstModel{
		script:    script,
		itemsList: list.New(nil, list.NewDefaultDelegate(), 80, 40),
	}
	mod.itemsList.Title = "Script AST"
	mod.itemsList.SetShowStatusBar(false)
	mod.itemsList.SetShowTitle(true)
	mod.itemsList.SetShowHelp(false)

	return mod
}

func (m *AstModel) Init() tea.Cmd {
	return tea.Batch(parseScript(m.script))
}

func (m *AstModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.itemsList.SetWidth(m.width - 4)
		m.itemsList.SetHeight(m.height - 4)
	case tea.KeyMsg:
		return m.updateKeys(msg)
	case astMsg:
		m.envelop = msg.envelop
		m.err = msg.err

		m.updateAST()
	}

	var cmd tea.Cmd
	m.itemsList, cmd = m.itemsList.Update(msg)

	return m, cmd
}

func (m *AstModel) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.NewView("")
	}

	// bf := lipgloss.NormalBorder()
	footer := lipgloss.NewStyle().Width(m.width).Render(m.script)

	// bb := lipgloss.NormalBorder()
	body := lipgloss.NewStyle().Height(m.height - 32).Width(m.width).Render(m.itemsList.View())

	screen := lipgloss.JoinVertical(
		lipgloss.Left,
		body,
		footer,
	)
	return tea.NewView(screen)
}

func (m *AstModel) updateAST() {
	var items []list.Item
	for _, n := range m.envelop.Root.Children {
		i := NodeItem{
			node: n,
		}
		items = append(items, i)
	}
	m.itemsList.SetItems(items)
	m.itemsList.Title = fmt.Sprintf("AST - %s", m.script)
}

func (m *AstModel) updateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	default:
		return m, nil
	}
}

func parseScript(file string) tea.Cmd {
	return func() tea.Msg {
		e, err := repr.InspectFile(file)
		return astMsg{
			envelop: e,
			err:     err,
		}
	}
}
