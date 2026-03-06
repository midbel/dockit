package ast

import (
	tea "charm.land/bubbletea/v2"
	"github.com/midbel/dockit/formula/repr"
)

type parseMsg struct {
	envelop *repr.Envelop
	err     error
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