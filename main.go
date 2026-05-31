package main

import (
	"fmt"

	app "Relay/app"
	screens "Relay/screens"

	tea "github.com/charmbracelet/bubbletea"
	// "github.com/charmbracelet/lipgloss"
)

func main() {
	p := tea.NewProgram(
		app.Model{
			Screen: screens.NewBootScreen(),
		},
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
