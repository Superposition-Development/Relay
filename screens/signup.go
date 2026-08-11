package screens

import (
	"Relay/app"
	_ "embed"
	"net/url"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SignupScreen struct {
	userID       textinput.Model
	password     textinput.Model
	username     textinput.Model
	serverAdress string
	focused      int

	width      int
	height     int
	loginError string
	useHTTP    bool
}

func NewSignupScreen(h, w int, serverAdress string, isHttp bool) SignupScreen {
	username := textinput.New()
	username.Placeholder = ""
	username.CharLimit = 32
	username.Width = 40
	username.Prompt = ""

	userID := textinput.New()
	userID.Placeholder = ""
	userID.CharLimit = 32
	userID.Width = 40
	userID.Prompt = ""

	password := textinput.New()
	password.Placeholder = ""
	password.CharLimit = 32
	password.Width = 40
	password.Prompt = ""
	password.EchoMode = textinput.EchoPassword

	username.Focus()

	return SignupScreen{
		userID:       userID,
		password:     password,
		username:     username,
		serverAdress: serverAdress,

		focused: 0,

		width:   w,
		height:  h,
		useHTTP: isHttp,
	}
}

func (m SignupScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m SignupScreen) Update(msg tea.Msg) (app.Screen, tea.Cmd) {

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

			if m.focused > 5 {
				m.focused = 0
			}

			m.updateFocus()

		case "shift+tab", "up":
			m.focused--

			if m.focused < 0 {
				m.focused = 5
			}

			m.updateFocus()

		case "enter":

			switch m.focused {

			case 0:

			case 1:

			case 2:

			case 3:
				payload := map[string]string{
					"username": m.username.Value(),
					"userID":   m.userID.Value(),
					"password": m.password.Value(),
					"pfp":      "",
				}
				protocol := "https"
				websocketProtocol := "wss"

				if m.useHTTP {
					protocol = "http"
					websocketProtocol = "ws"
				}

				server := url.URL{
					Scheme: protocol,
					Host:   m.serverAdress,
				}

				data, err := Signup(server.String(), payload)
				if err != nil {
					m.loginError = err.Error()
				}
				if data.RelayJWT != "" {
					app.SaveToken(data.RelayJWT)
					app.CurrentUserID = m.userID.Value()

					app.RegisterWebsocket(websocketProtocol+m.serverAdress+app.WebsocketEndpoint, data.RelayJWT, func(msg map[string]any) {
						app.WSChan <- msg
					})
					app.ServerURL.Host = m.serverAdress
					app.ServerURL.Scheme = protocol
					app.WebsocketURL.Host = m.serverAdress
					app.WebsocketURL.Scheme = websocketProtocol

					return m, func() tea.Msg {
						return app.ChangeScreenMsg{
							Screen: CreateChatScreen(m.height, m.width),
							Width:  m.width,
							Height: m.height,
						}
					}
				}

			case 4:
				return m, func() tea.Msg {
					return app.ChangeScreenMsg{
						Screen: NewStartScreen(m.height, m.width, false),
						Width:  m.width,
						Height: m.height,
					}
				}

			case 5:
				return m, func() tea.Msg {
					return app.ChangeScreenMsg{
						Screen: NewLoginScreen(m.height, m.width, m.serverAdress, m.useHTTP),
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

		m.username, cmd = m.username.Update(msg)

	case 1:

		m.userID, cmd = m.userID.Update(msg)

	case 2:

		m.password, cmd = m.password.Update(msg)
	}

	return m, cmd
}

func (m *SignupScreen) updateFocus() {
	m.username.Blur()
	m.userID.Blur()
	m.password.Blur()

	switch m.focused {

	case 0:

		m.username.Focus()

	case 1:

		m.userID.Focus()

	case 2:
		m.password.Focus()
	}
}

func (m SignupScreen) View() string {

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

	protocol := "https"

	if m.useHTTP {
		protocol = "http"

	}

	title := "Create Account for " + protocol + m.serverAdress

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

	usernameLabel := labelStyle.Render(
		"\nUsername:",
	)

	usernameField := inputStyle.Render(
		m.username.View(),
	)

	usernameRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		usernameLabel,
		usernameField,
	)

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
			m.focused == 3,
		),

		" ",

		button(
			"Cancel",
			12,
			m.focused == 4,
		),

		" ",

		button(
			"Login",
			14,
			m.focused == 5,
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
					usernameRow,

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

func Signup(url string, payload any) (LoginResponse, error) {

	var data LoginResponse

	err := app.POST(
		payload,
		"",
		url+app.SignupEndpoint,
		&data,
	)

	if err != nil {
		return LoginResponse{}, err
	}

	return data, nil
}
