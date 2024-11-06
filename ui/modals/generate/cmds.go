package generate

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Cancel struct{}

// EmitCancelGenerate cancel generation
func EmitCancel(cg Cancel) tea.Cmd {
	return func() tea.Msg {
		return cg
	}
}
