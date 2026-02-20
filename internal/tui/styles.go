package tui

import lipgloss "charm.land/lipgloss/v2"

var (
	accent = lipgloss.Color("#F97316")
	muted  = lipgloss.Color("#6B7280")
	white  = lipgloss.Color("#FAFAFA")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent)

	mutedStyle = lipgloss.NewStyle().
			Foreground(muted)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(muted)

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			Underline(true)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(muted)

	helpStyle = lipgloss.NewStyle().
			Foreground(muted).
			PaddingTop(1)
)

func statusStyle(status string) lipgloss.Style {
	switch status {
	case "COMPLETED":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E"))
	case "IN_PROGRESS", "QUEUED":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#EAB308"))
	case "FAILED":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	default:
		return lipgloss.NewStyle().Foreground(muted)
	}
}

func logLevelStyle(level string) lipgloss.Style {
	switch level {
	case "DEBUG":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	case "WARN":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#EAB308"))
	case "ERROR", "FATAL":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	}
}
