package ui

import (
	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Subtle    lipgloss.AdaptiveColor
	Highlight lipgloss.AdaptiveColor
	Special   lipgloss.AdaptiveColor
	Filtered  lipgloss.AdaptiveColor
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

func checkColors(colors []string) lipgloss.AdaptiveColor {
	if len(colors) == 0 {
		return lipgloss.AdaptiveColor{}
	}

	if len(colors) == 1 {
		return lipgloss.AdaptiveColor{Light: colors[0], Dark: colors[0]}
	}

	return lipgloss.AdaptiveColor{Light: colors[0], Dark: colors[1]}
}

func InitStyles() {
	Theme.Subtle = checkColors(api.AppConfig.Theme.Subtle)
	Theme.Highlight = checkColors(api.AppConfig.Theme.Highlight)
	Theme.Special = checkColors(api.AppConfig.Theme.Special)
	Theme.Filtered = checkColors(api.AppConfig.Theme.Filtered)

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
