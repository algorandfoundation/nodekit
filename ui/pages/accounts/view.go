package accounts

import (
	"github.com/algorandfoundation/hack-tui/ui/style"
)

func (m ViewModel) View() string {
	table := style.ApplyBorder(m.Width, m.Height, "8").Render(m.table.View())
	return style.WithNavigation(
		m.navigation,
		style.WithControls(
			m.controls,
			style.WithTitle(
				"Accounts",
				table,
			),
		),
	)
}
