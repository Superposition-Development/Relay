package screens

import (
	"Relay/app"
	_ "embed"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoginScreen struct {
	userID       textinput.Model
	password     textinput.Model
	serverAdress string
	focused      int
	isHttp       bool
	width        int
	height       int
	loginError   string
}

type LoginResponse struct {
	RelayJWT string `json:"RelayJWT"`
}

//go:embed logo.txt
var Logo string

func NewLoginScreen(h, w int, serverAdress string, isHttp bool) LoginScreen {
	userID := textinput.New()
	userID.Width = 40
	userID.Prompt = ""

	password := textinput.New()
	password.Width = 40
	password.Prompt = ""
	password.EchoMode = textinput.EchoPassword

	userID.Focus()

	return LoginScreen{
		userID:       userID,
		password:     password,
		focused:      0,
		height:       h,
		width:        w,
		serverAdress: serverAdress,
		isHttp:       isHttp,
	}
}

func (m LoginScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginScreen) Update(msg tea.Msg) (app.Screen, tea.Cmd) {

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

			if m.focused > 4 {
				m.focused = 0
			}

			m.updateFocus()

		case "shift+tab", "up":
			m.focused--

			if m.focused < 0 {
				m.focused = 4
			}

			m.updateFocus()

		case "enter":

			switch m.focused {

			case 0:
				// Username

			case 1:
				// Password

			case 2:
				payload := map[string]string{
					"userID":   m.userID.Value(),
					"password": m.password.Value(),
				}

				protocol := "https://"
				websocketProtocol := "wss://"

				if m.isHttp {
					protocol = "http://"
					websocketProtocol = "ws://"
				}

				data, err := Login(protocol+m.serverAdress, payload)
				if err != nil {
					m.loginError = err.Error()
				}
				if data.RelayJWT != "" {
					app.SaveToken(data.RelayJWT)
					app.CurrentUserID = m.userID.Value()
					app.SetCurrentServerAddress(m.serverAdress)
					app.RegisterWebsocket(websocketProtocol+m.serverAdress+app.WebsocketEndpoint, data.RelayJWT, func(msg map[string]any) {
						app.WSChan <- msg
					})
					return m, func() tea.Msg {
						return app.ChangeScreenMsg{
							Screen: CreateChatScreen(m.height, m.width),
							Width:  m.width,
							Height: m.height,
						}
					}
				}

			case 3:
				return m, func() tea.Msg {
					return app.ChangeScreenMsg{
						Screen: NewStartScreen(m.height, m.width, false),
						Width:  m.width,
						Height: m.height,
					}
				}

			case 4:
				return m, func() tea.Msg {
					return app.ChangeScreenMsg{
						Screen: NewSignupScreen(m.height, m.width, m.serverAdress, m.isHttp),
						Width:  m.width,
						Height: m.height,
					}
				}
			}
		}
	}

	var cmd tea.Cmd

	switch m.focused {

	case 0:

		m.userID, cmd = m.userID.Update(msg)

	case 1:

		m.password, cmd = m.password.Update(msg)
	}

	return m, cmd
}

func (m *LoginScreen) updateFocus() {
	m.userID.Blur()
	m.password.Blur()

	switch m.focused {

	case 0:

		m.userID.Focus()

	case 1:

		m.password.Focus()
	}
}

func (m LoginScreen) View() string {

	cream := lipgloss.Color("#F5B978")
	black := lipgloss.Color("#000000")
	selectedCream := lipgloss.Color("#ffdea5")

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555")).
		Width(82).
		Align(lipgloss.Right)

	titleStyle := lipgloss.NewStyle().
		Width(82).
		Foreground(cream)

	protocol := "https://"

	if m.isHttp {
		protocol = "http://"

	}

	title := "Login to " + protocol + m.serverAdress

	titleBar := lipgloss.JoinHorizontal(
		lipgloss.Top,

		titleStyle.Render(title),

		lipgloss.NewStyle().
			Width(58).
			Render(""),

		lipgloss.NewStyle().
			Foreground(cream).
			Render(),
	)

	logoStyle := lipgloss.NewStyle().
		Foreground(cream).
		Bold(true).
		Width(82).
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

	userLabel := labelStyle.Render(
		"\nUserID:",
	)

	userField := inputStyle.Render(
		m.userID.View(),
	)

	userRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		userLabel,
		userField,
	)

	passwordLabel := labelStyle.Render(
		"\nPassword:",
	)

	passwordField := inputStyle.Render(
		m.password.View(),
	)

	passwordRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		passwordLabel,
		passwordField,
	)

	button := func(
		text string,
		width int,
		focused bool,
	) string {

		style := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Border(lipgloss.NormalBorder()).
			BorderForeground(cream).
			Foreground(cream)

		if focused {

			style = style.
				BorderForeground(selectedCream)
		}

		return style.Render(text)
	}

	buttons := lipgloss.JoinHorizontal(

		lipgloss.Top,

		button(
			"OK",
			10,
			m.focused == 2,
		),

		" ",

		button(
			"Cancel",
			12,
			m.focused == 3,
		),

		" ",

		button(
			"Signup",
			14,
			m.focused == 4,
		),
	)

	errorMessage := ""

	if m.loginError != "" {
		errorMessage = errorStyle.Render(
			m.loginError,
		)
	}

	content := lipgloss.JoinVertical(

		lipgloss.Left,

		titleBar,

		lipgloss.NewStyle().
			Width(82).
			Border(lipgloss.NormalBorder()).
			BorderForeground(cream).
			Padding(1, 2).
			Render(

				lipgloss.JoinVertical(

					lipgloss.Left,

					logoStyle.Render(Logo),

					"",

					"",

					userRow,

					"",

					passwordRow,

					"",

					buttons,

					"",

					errorMessage,
				),
			),
	)

	window := lipgloss.NewStyle().
		Width(86).
		Height(21).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(cream).
		Foreground(cream).
		Background(black).
		Render(content)

	return lipgloss.Place(

		m.width,
		m.height,

		lipgloss.Center,
		lipgloss.Center,

		window,
	)
}

func Login(url string, payload any) (LoginResponse, error) {

	loginEndpoint := "/login"

	var data LoginResponse

	err := app.POST(
		payload,
		"",
		url+loginEndpoint,
		&data,
	)

	if err != nil {
		return LoginResponse{}, err
	}

	return data, nil
}
