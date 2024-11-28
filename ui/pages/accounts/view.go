package accounts

import (
	"github.com/algorandfoundation/hack-tui/ui/style"
)

func (m ViewModel) View() string {
	table := style.ApplyBorder(m.Width, m.Height, m.BorderColor).Render(m.table.View())
	ctls := m.Controls
	if m.Data.Status.LastRound < uint64(m.Data.Metrics.Window) {
		ctls = "( Insufficient Data )"
	}
	return style.WithNavigation(
		m.Navigation,
		style.WithControls(
			ctls,
			style.WithTitle(
				m.Title,
				table,
			),
		),
	)
}
