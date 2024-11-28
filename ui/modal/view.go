package modal

import (
	"github.com/algorandfoundation/hack-tui/ui/app"
	"github.com/algorandfoundation/hack-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {
	if !m.Open {
		return m.Parent
	}
	var render = ""
	switch m.Type {
	case app.InfoModal:
		render = m.infoModal.View()
	case app.TransactionModal:
		render = m.transactionModal.View()
	case app.ConfirmModal:
		render = m.confirmModal.View()
	case app.GenerateModal:
		render = m.generateModal.View()
	case app.ExceptionModal:
		render = m.exceptionModal.View()
	}
	width := lipgloss.Width(render) + 2
	height := lipgloss.Height(render)

	return style.WithOverlay(style.WithNavigation(
		m.controls,
		style.WithTitle(
			m.title,
			style.ApplyBorder(width, height, m.borderColor).
				PaddingRight(1).
				PaddingLeft(1).
				Render(render),
		),
	), m.Parent)
}
