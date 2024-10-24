package transaction

import "github.com/charmbracelet/lipgloss"

var qrStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("15")).
	Background(lipgloss.Color("0"))

var urlStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#2596be"))

var red = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9"))

var yellow = lipgloss.NewStyle().
	Foreground(lipgloss.Color("11"))

var Padding1 = lipgloss.NewStyle().Padding()
