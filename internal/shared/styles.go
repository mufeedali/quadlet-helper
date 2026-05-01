package shared

import (
	"os"
	"strconv"
)

// Style applies ANSI formatting. Honors NO_COLOR.
type Style struct {
	bold  bool
	color int // SGR color code; 0 = default
}

func (s Style) Render(text string) string {
	if os.Getenv("NO_COLOR") != "" || (!s.bold && s.color == 0) {
		return text
	}
	if s.bold && s.color != 0 {
		return "\x1b[1;" + strconv.Itoa(s.color) + "m" + text + "\x1b[0m"
	}
	if s.bold {
		return "\x1b[1m" + text + "\x1b[0m"
	}
	return "\x1b[" + strconv.Itoa(s.color) + "m" + text + "\x1b[0m"
}

var (
	TitleStyle    = Style{bold: true, color: 96} // Bold Bright Cyan
	SuccessStyle  = Style{color: 92}             // Bright Green
	ErrorStyle    = Style{color: 91}             // Bright Red
	WarningStyle  = Style{color: 93}             // Bright Yellow
	FilePathStyle = Style{color: 93}             // Bright Yellow

	CheckMark  = SuccessStyle.Render("✓")
	CrossMark  = ErrorStyle.Render("✗")
	InfoMark   = WarningStyle.Render("💡")
	FolderMark = TitleStyle.Render("📁")
)
