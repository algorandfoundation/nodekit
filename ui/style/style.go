package style

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
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
func WithControls(nav string, view string) string {
	if nav == "" {
		return view
	}
	controlWidth := lipgloss.Width(nav)
	if lipgloss.Width(view) >= controlWidth+4 {
		b, _, _, _, _ := Border.GetBorder()
		find := b.BottomLeft + strings.Repeat(b.Bottom, controlWidth+4)
		// TODO: allow other border colors, possibly just grab the last escape char
		return strings.Replace(view, find, b.BottomLeft+strings.Repeat(b.Bottom, 4)+"\u001B[0m"+nav+"\u001B[90m", 1)
	}
	return view
}
func WithNavigation(controls string, view string) string {
	if controls == "" {
		return view
	}
	controlWidth := lipgloss.Width(controls)
	if lipgloss.Width(view) >= controlWidth+4 {
		b, _, _, _, _ := Border.GetBorder()
		find := strings.Repeat(b.Bottom, controlWidth+4) + b.BottomRight
		// TODO: allow other border colors, possibly just grab the last escape char
		return strings.Replace(view, find, "\u001B[0m"+controls+"\u001B[90m"+strings.Repeat(b.Bottom, 4)+b.BottomRight, 1)
	}
	return view
}

const BANNER = `
   _____  .__                __________              
  /  _  \ |  |    ____   ____\______   \__ __  ____  
 /  /_\  \|  |   / ___\ /  _ \|       _/  |  \/    \ 
/    |    \  |__/ /_/  >  <_> )    |   \  |  /   |  \
\____|__  /____/\___  / \____/|____|_  /____/|___|  /
        \/     /_____/               \/           \/ 
`
