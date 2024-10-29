package style

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (
	Border = func() lipgloss.Style {
		return lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder())
	}()
	ApplyBorder = func(width int, height int, color string) lipgloss.Style {
		return Border.
			Width(width).
			Padding(0).
			Margin(0).
			Height(height).
			BorderForeground(lipgloss.Color(color))
	}

	Blue = func() lipgloss.Style {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	}()
	Cyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	Yellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	Green   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	Red     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	Magenta = lipgloss.NewStyle().
		Foreground(lipgloss.Color("5")).
		Render
	Purple = lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Render
	LightBlue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Render
)

func WithTitle(title string, view string) string {
	r := []rune(view)
	if lipgloss.Width(view) >= len(title)+4 {
		b, _, _, _, _ := Border.GetBorder()
		id := strings.IndexRune(view, []rune(b.Top)[0])
		start := string(r[0:id])
		return start + title + string(r[len(title)+id:])
	}
	return view
}
