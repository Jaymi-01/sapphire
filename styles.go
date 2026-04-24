package main

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	white  = lipgloss.Color("#FFFFFF")
	gray   = lipgloss.Color("#767676")
	red    = lipgloss.Color("#FF0000")
	green  = lipgloss.Color("#00FF00")
	blue   = lipgloss.Color("#0000FF")
	yellow = lipgloss.Color("#FFFF00")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			Background(blue).
			Padding(0, 1).
			MarginBottom(1)

	statsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(gray).
			Padding(0, 1).
			Width(30)

	mapStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(green).
			Padding(0, 1)

	logStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(gray).
			Height(5).
			Width(50).
			Padding(0, 1)

	playerStyle = lipgloss.NewStyle().Foreground(yellow).Bold(true)
	npcStyle    = lipgloss.NewStyle().Foreground(blue).Bold(true)
	itemStyle   = lipgloss.NewStyle().Foreground(green)
)
