package cmd

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00BFFF")) // DeepSkyBlue

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#32CD32")) // LimeGreen

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4500")) // OrangeRed

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")) // Gold

	filePathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")) // Yellow

	checkMark  = successStyle.Render("✓")
	crossMark  = errorStyle.Render("✗")
	infoMark   = warningStyle.Render("💡")
	folderMark = titleStyle.Render("📁")
)
