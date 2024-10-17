package ui

import "github.com/charmbracelet/lipgloss"

var (
	rounderBorder = func() lipgloss.Style {
		return lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder())
	}()
	topSections = func(width int) lipgloss.Style {
		return rounderBorder.
			Width(width - 2).
			Padding(0).
			Margin(0).
			Height(5).
			//BorderBackground(lipgloss.Color("4")).
			BorderForeground(lipgloss.Color("5"))
	}

	blue = func() lipgloss.Style {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	}()
	cyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	yellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	green   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
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
