package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View() string
}

type Model struct {
	Screen Screen
	width  int
	height int
}

type ChangeScreenMsg struct {
	Screen Screen
	Width  int
	Height int
}

func (m Model) Init() tea.Cmd {
	return m.Screen.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case ChangeScreenMsg:
		m.Screen = msg.Screen
		m.height = msg.Height
		m.width = msg.Width
		return m, m.Screen.Init()

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	nextScreen, cmd := m.Screen.Update(msg)
	m.Screen = nextScreen

	return m, cmd
}

func (m Model) View() string {
	return m.Screen.View()
}
