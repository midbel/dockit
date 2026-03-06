package ast

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/midbel/dockit/formula/repr"
	"github.com/midbel/dockit/studio/screen"
)

type astList struct {
	width  int
	height int
	list   list.Model

	envelop *repr.Envelop
	err     error
}

func NewList() screen.Screen {
	a := astList{
		list: list.New(nil, list.NewDefaultDelegate(), 0, 0),
	}
	return &a
}

func (a *astList) Init() tea.Cmd {
	return nil
}

func (a *astList) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	if pm, ok := msg.(parseMsg); ok {
		a.envelop = pm.envelop
		a.err = pm.err
		a.updateList(pm.envelop.Root)
	}
	var cmd tea.Cmd
	a.list, cmd = a.list.Update(msg)
	return a, cmd
}

func (a *astList) View() string {
	style := lipgloss.NewStyle()
	return style.Width(a.width).Render(a.list.View())
}

func (a *astList) Resize(w, h int) {
	a.width = w
	a.height = h
	a.list.SetSize(a.width, a.height)
}

func (a *astList) updateList(root *repr.Node) {
	if root == nil {
		return
	}
	var items []list.Item
	for _, n := range root.Children {
		i := astNodeItem{
			node: n,
		}
		items = append(items, i)
	}
	a.list.ResetSelected()
	a.list.SetItems(items)
	a.list.SetSize(a.width, a.height)
}

type astNodeItem struct {
	node *repr.Node
}

func (n astNodeItem) Title() string {
	return n.node.Name
}

func (n astNodeItem) Description() string {
	return n.node.Raw()
}

func (n astNodeItem) FilterValue() string {
	return n.node.Name
}
