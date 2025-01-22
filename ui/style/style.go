package style

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"regexp"
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
			Height(height).
			BorderForeground(lipgloss.Color(color))
	}
	MarketingBlue = lipgloss.NewStyle().Foreground(lipgloss.Color("#2d2df1")).Render
	MarketingTeal = lipgloss.NewStyle().Foreground(lipgloss.Color("#17CAC6")).Render
	Blue          = func() lipgloss.Style {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#12"))
	}()
	Cyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("#14"))
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
	Bold = lipgloss.NewStyle().
		Bold(true).
		Render
	BoldUnderline = lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Render
	Italics = lipgloss.NewStyle().
		Italic(true).
		Render
)

func WithHyperlink(text string, url string) string {
	return fmt.Sprintf("\033]8;;%s\a%s\033]8;;\a", url, text)
}
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

	lines := strings.Split(view, "\n")
	padLeft := 5

	if lipgloss.Width(view) >= controlWidth+4 {
		line := lines[len(lines)-1]
		leftEdge := padLeft
		lineLeft := ansi.Truncate(line, leftEdge, "")
		lineRight := TruncateLeft(line, leftEdge+lipgloss.Width(nav))
		lines[len(lines)-1] = lineLeft + nav + lineRight
	}
	return strings.Join(lines, "\n")
}
func WithNavigation(controls string, view string) string {
	if controls == "" {
		return view
	}

	padRight := 5
	controlWidth := lipgloss.Width(controls)

	lines := strings.Split(view, "\n")

	if lipgloss.Width(view) >= controlWidth+4 {
		line := lines[len(lines)-1]
		lineWidth := lipgloss.Width(line)
		leftEdge := lineWidth - (controlWidth + padRight)
		lineLeft := ansi.Truncate(line, leftEdge, "")
		lineRight := TruncateLeft(line, leftEdge+lipgloss.Width(controls))
		lines[len(lines)-1] = lineLeft + controls + lineRight
	}
	return strings.Join(lines, "\n")
}

// WithOverlay is the merging of two views
// Based on https://gist.github.com/Broderick-Westrope/b89b14770c09dda928c4a108f437b927
func WithOverlay(overlay string, view string) string {
	if overlay == "" {
		return view
	}

	bgLines := strings.Split(view, "\n")
	overlayLines := strings.Split(overlay, "\n")

	row := lipgloss.Height(view) / 2
	row -= lipgloss.Height(overlay) / 2
	col := lipgloss.Width(view) / 2
	col -= lipgloss.Width(overlay) / 2

	// Display the terminal resize
	if col < 0 || row < 0 {
		overlayText := "Resize terminal to see content"
		overlay = WithNavigation("( esc )", WithTitle("Screen Size",
			ApplyBorder(
				lipgloss.Width(overlayText)+4,
				lipgloss.Height(overlayText)+2,
				"1",
			).Padding(1).Render(overlayText),
		),
		)
		overlayLines = strings.Split(overlay, "\n")
		row = lipgloss.Height(view) / 2
		row -= lipgloss.Height(overlay) / 2
		col = lipgloss.Width(view) / 2
		col -= lipgloss.Width(overlay) / 2

		// If it is still too small, just return the text
		if col < 0 || row < 0 {
			return overlayText
		}
	}

	for i, overlayLine := range overlayLines {
		targetRow := i + row

		// Ensure the target row exists in the background lines
		for len(bgLines) <= targetRow {
			bgLines = append(bgLines, "")
		}

		bgLine := bgLines[targetRow]
		bgLineWidth := ansi.StringWidth(bgLine)

		if bgLineWidth < col {
			bgLine += strings.Repeat(" ", col-bgLineWidth) // Add padding
		}

		bgLeft := ansi.Truncate(bgLine, col, "")
		bgRight := TruncateLeft(bgLine, col+ansi.StringWidth(overlayLine))

		bgLines[targetRow] = bgLeft + overlayLine + bgRight
	}

	return strings.Join(bgLines, "\n")
}

// TruncateLeft removes characters from the beginning of a line, considering ANSI escape codes.
func TruncateLeft(line string, padding int) string {

	// This is genius, thank you https://gist.github.com/Broderick-Westrope/b89b14770c09dda928c4a108f437b927
	wrapped := strings.Split(ansi.Hardwrap(line, padding, true), "\n")
	if len(wrapped) == 1 {
		return ""
	}

	var ansiStyle string
	// Regular expression to match ANSI escape codes.
	ansiStyles := regexp.MustCompile(`\x1b[[\d;]*m`).FindAllString(wrapped[0], -1)
	if len(ansiStyles) > 0 {
		// Pick the last style found
		ansiStyle = ansiStyles[len(ansiStyles)-1]
	}

	return ansiStyle + strings.Join(wrapped[1:], "")
}

const BANNER_NODE = `
███    ██  ██████  ██████  ███████ 
████   ██ ██    ██ ██   ██ ██      
██ ██  ██ ██    ██ ██   ██ █████   
██  ██ ██ ██    ██ ██   ██ ██      
██   ████  ██████  ██████  ███████ `
const BANNER_KIT = `
██╗  ██╗██╗████████╗
██║ ██╔╝██║╚══██╔══╝
█████╔╝ ██║   ██║   
██╔═██╗ ██║   ██║   
██║  ██╗██║   ██║   
╚═╝  ╚═╝╚═╝   ╚═╝`

var BANNER string

func init() {
	var res string
	lines := strings.Split(BANNER_KIT, "\n")
	for i, line := range strings.Split(BANNER_NODE, "\n") {
		res += MarketingBlue(line) + MarketingTeal(lines[i]) + "\n"
	}
	BANNER = res
}
