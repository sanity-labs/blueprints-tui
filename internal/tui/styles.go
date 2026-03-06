package tui

import (
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

type styles struct {
	headerBox    lipgloss.Style
	headerTitle  lipgloss.Style
	headerHint   lipgloss.Style
	headerValue  lipgloss.Style
	title        lipgloss.Style
	muted        lipgloss.Style
	sectionHead  lipgloss.Style
	tabActive    lipgloss.Style
	tabInactive  lipgloss.Style
	tabUnderline lipgloss.Style
	keycap       lipgloss.Style
	help         lipgloss.Style

	statusCompleted  lipgloss.Style
	statusFailed     lipgloss.Style
	statusInProgress lipgloss.Style
	statusDefault    lipgloss.Style

	logDebug   lipgloss.Style
	logWarn    lipgloss.Style
	logError   lipgloss.Style
	logFatal   lipgloss.Style
	logDefault lipgloss.Style

	table table.Styles
}

func newStyles(isDark bool) styles {
	lightDark := lipgloss.LightDark(isDark)

	accent := lightDark(lipgloss.Color("63"), lipgloss.Color("69"))
	fg := lightDark(lipgloss.Color("235"), lipgloss.Color("252"))
	muted := lightDark(lipgloss.Color("247"), lipgloss.Color("243"))
	faint := lightDark(lipgloss.Color("250"), lipgloss.Color("238"))
	inverse := lightDark(lipgloss.Color("255"), lipgloss.Color("235"))

	green := lightDark(lipgloss.Color("32"), lipgloss.Color("75"))
	red := lightDark(lipgloss.Color("160"), lipgloss.Color("203"))
	yellow := lightDark(lipgloss.Color("172"), lipgloss.Color("214"))
	pink := lightDark(lipgloss.Color("162"), lipgloss.Color("205"))

	var s styles

	s.headerBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(faint).
		Padding(0, 1)

	s.headerTitle = lipgloss.NewStyle().
		Foreground(inverse).
		Background(accent).
		Padding(0, 1).
		Bold(true)

	s.headerHint = lipgloss.NewStyle().
		Foreground(muted)

	s.headerValue = lipgloss.NewStyle().
		Foreground(fg)

	s.title = lipgloss.NewStyle().
		Bold(true).
		Foreground(accent)

	s.muted = lipgloss.NewStyle().
		Foreground(muted)

	s.sectionHead = lipgloss.NewStyle().
		Bold(true).
		Foreground(fg).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(faint)

	s.tabActive = lipgloss.NewStyle().
		Foreground(inverse).
		Background(accent).
		Padding(0, 1).
		Bold(true)

	s.tabInactive = lipgloss.NewStyle().
		Foreground(muted).
		Padding(0, 1)

	s.tabUnderline = lipgloss.NewStyle().
		Foreground(accent)

	s.keycap = lipgloss.NewStyle().
		Foreground(lightDark(lipgloss.Color("235"), lipgloss.Color("235"))).
		Background(lightDark(lipgloss.Color("254"), lipgloss.Color("250"))).
		Padding(0, 1).
		Bold(true)

	s.help = lipgloss.NewStyle().
		Foreground(muted).
		PaddingTop(1)

	// Status colors — blue for success to avoid red/green colorblindness issues
	s.statusCompleted = lipgloss.NewStyle().Foreground(green)
	s.statusFailed = lipgloss.NewStyle().Foreground(red)
	s.statusInProgress = lipgloss.NewStyle().Foreground(yellow)
	s.statusDefault = lipgloss.NewStyle().Foreground(muted)

	// Log level colors
	s.logDebug = lipgloss.NewStyle().Foreground(muted)
	s.logWarn = lipgloss.NewStyle().Foreground(yellow)
	s.logError = lipgloss.NewStyle().Foreground(red)
	s.logFatal = lipgloss.NewStyle().Foreground(pink)
	s.logDefault = lipgloss.NewStyle().Foreground(fg)

	// Table styles
	ts := table.DefaultStyles()
	ts.Header = ts.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(faint).
		BorderBottom(true).
		Bold(false)
	ts.Selected = ts.Selected.
		Foreground(inverse).
		Background(accent).
		Bold(false)
	s.table = ts

	return s
}

func (s styles) statusStyle(status string) lipgloss.Style {
	switch strings.ToUpper(status) {
	case "COMPLETED", "SUCCESS":
		return s.statusCompleted
	case "IN_PROGRESS", "IN PROGRESS", "QUEUED":
		return s.statusInProgress
	case "FAILED":
		return s.statusFailed
	default:
		return s.statusDefault
	}
}

// statusIndicator returns a colorblind-friendly shape+color indicator.
// Shape encodes state: ● = success, ■ = failure, ◆ = in progress, ○ = default.
func (s styles) statusIndicator(status string) string {
	switch strings.ToUpper(status) {
	case "COMPLETED", "SUCCESS":
		return s.statusCompleted.Render("●")
	case "FAILED":
		return s.statusFailed.Render("■")
	case "IN_PROGRESS", "IN PROGRESS", "QUEUED":
		return s.statusInProgress.Render("◆")
	default:
		return s.statusDefault.Render("○")
	}
}

func (s styles) logLevelStyle(level string) lipgloss.Style {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return s.logDebug
	case "WARN", "WARNING":
		return s.logWarn
	case "ERROR":
		return s.logError
	case "FATAL":
		return s.logFatal
	default:
		return s.logDefault
	}
}
