package ast

import (
	tea "charm.land/bubbletea/v2"
	"github.com/midbel/dockit/formula/repr"
)

type parseMsg struct {
	envelop *repr.Envelop
	err     error
}

type AstApp struct {
	width  int
	height int
	script string

	envelop *repr.Envelop
	err     error
}

func App(script string) *AstApp {
	mod := &AstApp{
		script: script,
	}

	return mod
}

func (m *AstApp) Init() tea.Cmd {
	return tea.Batch(parseScript(m.script))
}

func (m *AstApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		return m.updateKeys(msg)
	case parseMsg:
	}
	return m, nil
}

func (m *AstApp) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.NewView("")
	}
	return tea.NewView("")
}

func (m *AstApp) updateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return parseMsg{
			envelop: e,
			err: err,
		}
	}
}
