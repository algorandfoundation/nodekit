package overlay

import (
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/style"
)

// View renders the current modal's UI based on its type and state, or returns the parent content if the modal is closed.
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

	return style.WithOverlay(render, m.Parent)
}
