package screens

import (
	"Relay/app"
	"strconv"

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

type ChatScreen struct {
	width  int
	height int
}

func CreateModel(h, w int) ChatScreen {

	return ChatScreen{width: w,
		height: h}
}

func (m ChatScreen) Init() tea.Cmd {
	return nil
}

func (m ChatScreen) Update(msg tea.Msg) (app.Screen, tea.Cmd) {
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

func (m ChatScreen) View() string {
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

func (m ChatScreen) renderServerBar() string {

	return ServerBarStyle.
		Height(m.height - 1).
		Render(strconv.Itoa(m.height))
}

func (m ChatScreen) renderChannelBar() string {

	return channelBarStyle.
		Height(m.height - 1).
		// Render(out.String())
		Render()
}
