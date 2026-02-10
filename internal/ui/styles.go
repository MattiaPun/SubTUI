package ui

import (
	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Subtle    lipgloss.AdaptiveColor
	Highlight lipgloss.AdaptiveColor
	Special   lipgloss.AdaptiveColor
}

var Theme Styles

var (
	borderStyle       lipgloss.Style
	activeBorderStyle lipgloss.Style
	loginBoxStyle     lipgloss.Style
	loginHeaderStyle  lipgloss.Style
	loginHelpStyle    lipgloss.Style
	popupStyle        lipgloss.Style
)

func InitStyles() {
	Theme.Subtle = lipgloss.AdaptiveColor{Light: api.AppConfig.Theme.Subtle[0], Dark: api.AppConfig.Theme.Subtle[1]}
	Theme.Highlight = lipgloss.AdaptiveColor{Light: api.AppConfig.Theme.Highlight[0], Dark: api.AppConfig.Theme.Highlight[1]}
	Theme.Special = lipgloss.AdaptiveColor{Light: api.AppConfig.Theme.Special[0], Dark: api.AppConfig.Theme.Special[1]}

	// Global Borders
	borderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.Subtle)

	// Focused Border (Brighter)
	activeBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.Highlight)

	loginBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.Highlight).
		Padding(1, 4).
		Align(lipgloss.Center)

	// The "Welcome" header
	loginHeaderStyle = lipgloss.NewStyle().
		Foreground(Theme.Special).
		Bold(true).
		MarginBottom(1)

	// The footer instruction
	loginHelpStyle = lipgloss.NewStyle().
		Foreground(Theme.Subtle).
		MarginTop(2)

	// The popup
	popupStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.Highlight).
		Padding(1, 2)

}
