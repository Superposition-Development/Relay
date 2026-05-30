package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	HeaderStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#444")).
		Foreground(lipgloss.Color("#FFF"))
)

type model struct {
	width  int
	height int
}

func createModel() model {

	return model{}
}

func (m model) Init() tea.Cmd {
	return nil //textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		}
	}
	return m, cmd
}

func (m model) View() string {
	header := HeaderStyle.Width(m.width).Render("Relay")

	body := "Main content here"

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
	)
}

func main() {
	p := tea.NewProgram(
		createModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
