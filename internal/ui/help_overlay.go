package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func NewHelpModel() HelpModel { return HelpModel{} }

func (m HelpModel) Init() tea.Cmd                           { return nil }
func (m HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m HelpModel) View() string                            { return helpViewContent() }

func (m BackgroundWrapper) Init() tea.Cmd                           { return nil }
func (m BackgroundWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m BackgroundWrapper) View() string                            { return m.RenderedView }
