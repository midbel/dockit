package screen

import (
	tea "charm.land/bubbletea/v2"
)

type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View() string
	Resize(int, int)
}

type FocusableScreen interface {
	Screen
	Focus()
	Blur()
}
