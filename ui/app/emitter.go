package app

import tea "github.com/charmbracelet/bubbletea"

type Outside chan tea.Msg

func NewOutside() Outside {
	return make(chan tea.Msg)
}

func (o Outside) Emit(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		o <- msg
		return nil
	}
}
