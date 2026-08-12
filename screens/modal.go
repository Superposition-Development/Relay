package screens

import (
	"Relay/app"
	"fmt"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var inputStyle = lipgloss.NewStyle().
	Width(40).
	Height(1).
	Border(lipgloss.NormalBorder()).
	BorderForeground(cream).
	Foreground(cream)

type Modal interface {
	Update(msg tea.Msg) (Modal, tea.Cmd, any)
	View() string
}

func (m ChatScreen) renderModalContent() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(cream).Render("Help & Keyboard Shortcuts")
	content := fmt.Sprintf(
		"%s\n\n"+
			"• Tab        : Switch focus between panels\n"+
			"• Enter      : Select menu / Send message\n"+
			"• Ctrl+W     : Toggle write mode\n"+
			"• Ctrl+S     : Send message (write mode)\n"+
			"• Ctrl+O     : Toggle this overlay\n"+
			"• Esc        : Exit active menu / close modal\n\n"+
			"%s",
		title,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#767676")).Render("Press Esc or 'q' to close"),
	)
	return modalBoxStyle.Render(content)
}

type JoinServerModal struct {
	input textinput.Model
}

func NewJoinServerModal() *JoinServerModal {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Focus()
	return &JoinServerModal{input: ti}
}

func (m *JoinServerModal) Update(msg tea.Msg) (Modal, tea.Cmd, any) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			return nil, nil, nil
		case "enter":

			JoinServer(m.input.Value())
			return nil, nil, m.input.Value()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd, nil
}

func (m *JoinServerModal) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(cream).Render("Join Server")
	idRow := lipgloss.JoinHorizontal(lipgloss.Top, "\nServerID: ", inputStyle.Render(m.input.View()))

	content := fmt.Sprintf("%s\n\n%s\n\n%s",
		title,
		idRow,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#767676")).Render("Press Enter to join • Esc to close"),
	)
	return modalBoxStyle.Render(content)
}

type CreateChannelModal struct {
	channelName textinput.Model
	focused     int
	isText      bool //false means voice
}

func NewCreateChannelModal() *CreateChannelModal {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Focus()
	return &CreateChannelModal{channelName: ti, focused: 0, isText: true}
}

func (m *CreateChannelModal) Update(msg tea.Msg) (Modal, tea.Cmd, any) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			return nil, nil, nil

		case "tab", "shift+tab":
			if m.focused == 0 {
				m.focused = 1
				m.channelName.Blur()
			} else {
				m.focused = 0
				m.channelName.Focus()
			}
			return m, nil, nil

		case "enter":
			if m.focused == 1 {
				m.isText = !m.isText
				return m, nil, nil
			}

			channelType := "voice"
			if m.isText {
				channelType = "text"
			}
			CreateChannel(fmt.Sprintf("%v", app.CurrentServerID), m.channelName.Value(), channelType)
			return nil, nil, m.channelName.Value()
		}
	}

	var cmd tea.Cmd
	if m.focused == 0 {
		m.channelName, cmd = m.channelName.Update(msg)
	}
	return m, cmd, nil
}

func (m *CreateChannelModal) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(cream).Render("Create Channel")
	renderButton := func(text string, width int, focused bool) string {
		style := lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Border(lipgloss.NormalBorder()).
			BorderForeground(cream).
			Foreground(cream)

		if focused {
			style = style.
				BorderForeground(selectedCream).
				Foreground(selectedCream).
				Bold(true)
		}

		return style.Render(text)
	}

	buttonLabel := "Voice Channel"
	if m.isText {
		buttonLabel = "Text Channel"
	}

	textStyle := inputStyle
	if m.focused == 0 {
		textStyle = textStyle.BorderForeground(selectedCream).
			Foreground(selectedCream).
			Bold(true)
	}

	button := renderButton(buttonLabel, 10, m.focused == 1)
	idRow := lipgloss.JoinHorizontal(lipgloss.Top, "\nName: ", textStyle.Render(m.channelName.View()))

	content := fmt.Sprintf("%s\n\n%s\n%s\n%s",
		title,
		idRow,
		button,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#767676")).Render("Press Enter to join • Esc to close"),
	)
	return modalBoxStyle.Render(content)
}

func overlayModal(baseView, modalView string, totalWidth, totalHeight int) string {
	baseView = dimStyle.Render(baseView)

	baseLines := strings.Split(baseView, "\n")
	modalLines := strings.Split(modalView, "\n")

	modalWidth := lipgloss.Width(modalView)
	modalHeight := len(modalLines)

	startX := (totalWidth - modalWidth) / 2
	startY := (totalHeight - modalHeight) / 2

	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	for i, mLine := range modalLines {
		targetY := startY + i
		if targetY >= len(baseLines) {
			break
		}

		bLine := baseLines[targetY]
		left := truncateANSI(bLine, startX)
		right := skipANSI(bLine, startX+modalWidth)

		baseLines[targetY] = left + mLine + right
	}

	return strings.Join(baseLines, "\n")
}

func truncateANSI(s string, width int) string {
	var result strings.Builder
	currentWidth := 0
	inAnsi := false

	for _, r := range s {
		if r == '\x1b' {
			inAnsi = true
		}
		if inAnsi {
			result.WriteRune(r)
			if r == 'm' {
				inAnsi = false
			}
			continue
		}
		if currentWidth >= width {
			break
		}
		rw := lipgloss.Width(string(r))
		if currentWidth+rw > width {
			break
		}
		result.WriteRune(r)
		currentWidth += rw
	}
	return result.String()
}

func skipANSI(s string, offset int) string {
	var result strings.Builder
	currentWidth := 0
	inAnsi := false
	pastOffset := false

	for _, r := range s {
		if r == '\x1b' {
			inAnsi = true
		}
		if inAnsi {
			if pastOffset {
				result.WriteRune(r)
			}
			if r == 'm' {
				inAnsi = false
			}
			continue
		}

		rw := lipgloss.Width(string(r))
		if currentWidth >= offset {
			pastOffset = true
			result.WriteRune(r)
		} else {
			currentWidth += rw
		}
	}
	return result.String()
}

func JoinServer(serverID string) {

	var response any

	payload := map[string]string{
		"serverID": serverID,
	}

	token, err := app.LoadToken()
	if err != nil {

	}

	url := url.URL{
		Scheme: app.ServerURL.Scheme,
		Host:   app.ServerURL.Host,
		Path:   app.JoinServerEndpoint,
	}

	err = app.POST(
		payload,
		token,
		url.String(),
		&response,
	)

}

func CreateChannel(serverID string, channelName string, channelType string) {
	var response any

	payload := map[string]string{
		"serverID": serverID,
		"name":     channelName,
		"type":     channelType,
	}

	token, err := app.LoadToken()
	if err != nil {

	}

	url := url.URL{
		Scheme: app.ServerURL.Scheme,
		Host:   app.ServerURL.Host,
		Path:   app.CreateChannelEndpoint,
	}

	err = app.POST(
		payload,
		token,
		url.String(),
		&response,
	)

}
