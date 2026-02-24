package tui

import "github.com/charmbracelet/lipgloss"

var (
	accent = lipgloss.Color("#3B82F6")
	muted  = lipgloss.Color("#6B7280")
	white  = lipgloss.Color("#FAFAFA")
	dark   = lipgloss.Color("#0F172A")

	headerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(muted).
			Padding(0, 1)

	headerTitleStyle = lipgloss.NewStyle().
				Foreground(dark).
				Background(accent).
				Padding(0, 1).
				Bold(true)

	headerHintStyle = lipgloss.NewStyle().
				Foreground(muted)

	headerValueStyle = lipgloss.NewStyle().
				Foreground(white)

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
			Foreground(dark).
			Background(accent).
			Padding(0, 1).
			Bold(true)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(muted).
				Padding(0, 1)

	tabUnderlineStyle = lipgloss.NewStyle().
				Foreground(accent)

	keycapStyle = lipgloss.NewStyle().
			Foreground(dark).
			Background(lipgloss.Color("#E5E7EB")).
			Padding(0, 1).
			Bold(true)

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
