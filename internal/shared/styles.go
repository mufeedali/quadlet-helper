package shared

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00BFFF")) // DeepSkyBlue

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#32CD32")) // LimeGreen

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4500")) // OrangeRed

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")) // Gold

	FilePathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")) // Yellow

	CheckMark  = SuccessStyle.Render("✓")
	CrossMark  = ErrorStyle.Render("✗")
	InfoMark   = WarningStyle.Render("💡")
	FolderMark = TitleStyle.Render("📁")
)
