package screens

import (
	"Relay/app"
	_ "embed"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StartScreen struct {
	connect textinput.Model
	focused int

	width  int
	height int
}

func NewStartScreen(h, w int) StartScreen {
	connect := textinput.New()
	connect.Width = 40
	connect.Prompt = ""

	connect.Focus()

	return StartScreen{
		connect: connect,
		focused: 0,
		height:  h,
		width:   w,
	}
}

func initialStartScreen() StartScreen {
	connect := textinput.New()
	connect.Placeholder = ""
	connect.CharLimit = 32
	connect.Width = 40

	connect.Prompt = ""

	connect.Focus()

	return StartScreen{
		connect: connect,

		focused: 0,

		width:  100,
		height: 30,
	}
}

func (m StartScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m StartScreen) Update(msg tea.Msg) (app.Screen, tea.Cmd) {

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

			if m.focused > 3 {
				m.focused = 0
			}

			m.updateFocus()

		case "shift+tab", "up":
			m.focused--

			if m.focused < 0 {
				m.focused = 3
			}

			m.updateFocus()

		case "enter":

			switch m.focused {

			case 0:
				// Username

			case 1:
				// Password

			case 2:

			case 3:
				//cancel
			}
		}
	}

	var cmd tea.Cmd

	switch m.focused {

	case 0:

		m.connect, cmd = m.connect.Update(msg)

	}
	return m, cmd

}

func (m *StartScreen) updateFocus() {
	m.connect.Blur()

	switch m.focused {

	case 0:

		m.connect.Focus()

	}
}

func (m StartScreen) View() string {

	cream := lipgloss.Color("#F5B978")
	black := lipgloss.Color("#000000")
	selectedCream := lipgloss.Color("#ffdea5")

	titleStyle := lipgloss.NewStyle().
		Width(82).
		Foreground(cream)

	title := "Relay CLI Edition"

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
		"\nConnect to:",
	)

	userField := inputStyle.Render(
		m.connect.View(),
	)

	protocol := lipgloss.NewStyle().
		Foreground(cream).
		Render("\nhttp://")

	inputWithProtocol := lipgloss.JoinHorizontal(
		lipgloss.Top,
		protocol,
		userField,
	)

	userRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		userLabel,
		inputWithProtocol,
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
	)

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

					buttons,
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
