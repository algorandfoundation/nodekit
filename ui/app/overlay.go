package app

import tea "github.com/charmbracelet/bubbletea"

type OverlayEventType string

const (
	OverlayEventClose  OverlayEventType = "close"
	OverlayEventCancel OverlayEventType = "cancel"
)

func EmitCloseOverlay() tea.Cmd {
	return func() tea.Msg {
		return OverlayEventClose
	}
}

func EmitCancelOverlay() tea.Cmd {
	return func() tea.Msg {
		return OverlayEventCancel
	}
}
