package generate

import (
	"fmt"
	"strings"

	"github.com/algorandfoundation/hack-tui/ui/style"
)

func (m ViewModel) View() string {
	var b strings.Builder

	for i := range m.Inputs {
		b.WriteString(m.Inputs[i].View())
		if i < len(m.Inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.Inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	render := style.ApplyBorder(m.Width, m.Height, "8").Render(b.String())
	return style.WithControls(
		m.controls,
		style.WithTitle(
			"Generate",
			render,
		),
	)
}
