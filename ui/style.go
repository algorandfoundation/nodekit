package ui

import "github.com/charmbracelet/lipgloss"

var (
	Magenta = lipgloss.NewStyle().
		Foreground(lipgloss.Color("5")).
		Render
	Muted = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render
	Purple = lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Render
	LightBlue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Render
)
