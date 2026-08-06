package screens

import (
	"strings"

	"Relay/app"
	_ "embed"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//go:embed serverart.txt
var serverArt string

type CreateServerScreen struct {
	serverName textinput.Model
	focused    int
	isHttp     bool
	width      int
	height     int
	loginError string
}

func NewCreateServerScreen(h, w int) CreateServerScreen {
	serverName := textinput.New()
	serverName.Width = 40
	serverName.Prompt = ""

	serverName.Focus()

	return CreateServerScreen{
		serverName: serverName,
		focused:    0,
		height:     h,
		width:      w,
	}
}

func (m CreateServerScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreateServerScreen) Update(msg tea.Msg) (app.Screen, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab", "down":
			m.focused++
			if m.focused > 2 {
				m.focused = 0
			}
			m.updateFocus()

		case "shift+tab", "up":
			m.focused--
			if m.focused < 0 {
				m.focused = 2
			}
			m.updateFocus()

		case "enter":
			if m.focused == 0 || m.focused == 1 {
				payload := map[string]string{
					"name": m.serverName.Value(),
					"pfp":  "",
				}
				CreateServer(payload)
			}
			return m, func() tea.Msg {
				return app.ChangeScreenMsg{
					Screen: CreateChatScreen(m.height, m.width),
					Width:  m.width,
					Height: m.height,
				}
			}

		}
	}

	var cmd tea.Cmd

	switch m.focused {
	case 0:
		m.serverName, cmd = m.serverName.Update(msg)
	}

	return m, cmd
}

func (m *CreateServerScreen) updateFocus() {
	m.serverName.Blur()

	switch m.focused {
	case 0:
		m.serverName.Focus()
	}
}

func (m CreateServerScreen) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	cream := lipgloss.Color("#F5B978")
	selectedCream := lipgloss.Color("#ffdea5")
	borderStyle := lipgloss.NewStyle().Foreground(cream)

	logoStyle := lipgloss.NewStyle().
		Foreground(cream).
		Bold(true).
		Align(lipgloss.Center)

	labelStyle := lipgloss.NewStyle().
		Width(14).
		Height(1).
		Foreground(cream).
		AlignVertical(lipgloss.Center)

	inputStyle := lipgloss.NewStyle().
		Width(40).
		Height(1).
		Border(lipgloss.NormalBorder()).
		BorderForeground(cream).
		Foreground(cream)

	userRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		labelStyle.Render("\nServer Name:"),
		inputStyle.Render(m.serverName.View()),
	)

	button := func(text string, width int, focused bool) string {
		style := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Border(lipgloss.NormalBorder()).
			BorderForeground(cream).
			Foreground(cream)

		if focused {
			style = style.BorderForeground(selectedCream)
		}

		return style.Render(text)
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Top,
		button("OK", 10, m.focused == 1),
		" ",
		button("Cancel", 12, m.focused == 2))

	errorMessage := ""
	if m.loginError != "" {
		errorMessage = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Width(58).
			Align(lipgloss.Center).
			Render(m.loginError)
	}

	formCard := lipgloss.NewStyle().
		Width(64).
		Border(lipgloss.NormalBorder()).
		BorderForeground(cream).
		Padding(1, 2).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				logoStyle.Render(serverArt),
				"",
				userRow,
				"",
				buttons,
				"",
				errorMessage,
			),
		)

	title := "Create Server"
	headerWidth := clamp(m.width-len(title)-5, 0, m.width)
	panelHeight := clamp(m.height-5, 4, m.height)

	top := borderStyle.Render("┌─ " + title + " " + strings.Repeat("─", headerWidth) + "┐")

	addr := app.GetCurrentServerAddress()
	userID := app.CurrentUserID
	spacing := clamp(m.width-len(addr)-len(userID)-2, 1, m.width)
	middle := borderStyle.Render("│" + addr + strings.Repeat(" ", spacing) + userID + "│")

	separator := borderStyle.Render("├" + strings.Repeat("─", m.width-2) + "┤")
	bottom := borderStyle.Render("└" + strings.Repeat("─", m.width-2) + "┘")

	innerWidth := m.width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}

	centeredBody := lipgloss.Place(
		innerWidth,
		panelHeight,
		lipgloss.Center,
		lipgloss.Center,
		formCard,
	)

	bodyLines := strings.Split(centeredBody, "\n")
	var framedBodyLines []string

	for _, line := range bodyLines {
		framedBodyLines = append(framedBodyLines, borderStyle.Render("│")+line+borderStyle.Render("│"))
	}

	var formattedContent []string
	formattedContent = append(formattedContent, top, middle, separator)
	formattedContent = append(formattedContent, framedBodyLines...)
	formattedContent = append(formattedContent, bottom)

	return strings.Join(formattedContent, "\n")
}

func CreateServer(payload any) {

	var response any

	token, err := app.LoadToken()
	if err != nil {

	}

	protocol := "http://"
	if app.IsServerSecure {
		protocol = "https://"
	}

	err = app.POST(
		payload,
		token,
		protocol+app.GetCurrentServerAddress()+app.CreateServerEndpoint,
		&response,
	)

	if err != nil {

	}
}
