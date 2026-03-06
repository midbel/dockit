package ast

import (
	tea "charm.land/bubbletea/v2"
	"github.com/midbel/dockit/formula/repr"
	"github.com/midbel/dockit/studio/screen"
)

type AstApp struct {
	width  int
	height int
	script string

	envelop *repr.Envelop
	err     error

	screen screen.Screen
}

func App(script string) *AstApp {
	a := &AstApp{
		script: script,
		screen: NewList(),
	}
	return a
}

func (m *AstApp) Init() tea.Cmd {
	return tea.Batch(parseScript(m.script))
}

func (m *AstApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.screen.Resize(m.width, m.height)
	case tea.KeyMsg:
		cmd = m.updateKeys(msg)
	case parseMsg:
		m.envelop = msg.envelop
		m.err = msg.err
	}
	var scmd tea.Cmd
	m.screen, scmd = m.screen.Update(msg)
	return m, tea.Batch(cmd, scmd)
}

func (m *AstApp) View() tea.View {
	view := tea.NewView("")
	view.AltScreen = true
	if m.width == 0 || m.height == 0 {
		return view
	}
	view.SetContent(m.screen.View())
	return view
}

func (m *AstApp) updateKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "q", "ctrl+c":
		return tea.Quit
	default:
		return nil
	}
}
