package generate

import (
	"fmt"
	"github.com/algorandfoundation/hack-tui/ui/style"
	"strings"
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

	return style.WithTitle("Generate", style.ApplyBorder(m.Width-3, m.Height-1, "8").Render(b.String()))
}
