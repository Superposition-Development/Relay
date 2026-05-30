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

	ServerBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#121212")).
			Width(10)

	channelBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E1E1E")).
			Width(28)
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

	serverBody := m.renderServerBar()
	channelBody := m.renderChannelBar()
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		serverBody,
		channelBody,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
	)
}

func (m model) renderServerBar() string {
	// var out strings.Builder

	// out.WriteString("\n")

	// out.WriteString(
	// 	lipgloss.NewStyle().
	// 		Foreground(lipgloss.Color("#FFFFFF")).
	// 		Bold(true).
	// 		Render(" SRD "),
	// )

	// out.WriteString("\n\n")

	// out.WriteString("10A 250VAC\n")
	// out.WriteString("10A 125VAC\n")
	// out.WriteString("10A  30VDC\n")
	// out.WriteString("SRD-05VDC\n")

	// out.WriteString("\n")

	// for _, s := range m.servers {
	// 	box := lipgloss.NewStyle().
	// 		Width(6).
	// 		Height(3).
	// 		Align(lipgloss.Center, lipgloss.Center).
	// 		Background(lipgloss.Color("#2F6DB2")).
	// 		Foreground(lipgloss.Color("#FFFFFF")).
	// 		MarginTop(1).
	// 		Render(s.Name)

	// 	out.WriteString(box)
	// 	out.WriteString("\n")
	// }

	return ServerBarStyle.
		Height(m.height - 1).
		Render()
}

func (m model) renderChannelBar() string {
	// var out strings.Builder

	// serverName := serverNameStyle.
	// 	Width(28).
	// 	Render("ServerName of infinite doom")

	// out.WriteString(serverName)
	// out.WriteString("\n")

	// for i, ch := range m.channels {
	// 	style := channelStyle

	// 	if i == m.selectedChannel {
	// 		style = selectedChannelStyle
	// 	}

	// 	out.WriteString(
	// 		style.Width(28).Render("# " + ch.Name),
	// 	)

	// 	out.WriteString("\n")
	// }

	return channelBarStyle.
		Height(m.height - 1).
		// Render(out.String())
		Render()
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
