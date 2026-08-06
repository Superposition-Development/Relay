package screens

import (
	"Relay/app"
	"fmt"
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
	input textinput.Model
}

func NewCreateChannelModal() *CreateChannelModal {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Focus()
	return &CreateChannelModal{input: ti}
}

func (m *CreateChannelModal) Update(msg tea.Msg) (Modal, tea.Cmd, any) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			return nil, nil, nil
		case "enter":

			CreateChannel(fmt.Sprintf("%v", app.CurrentServerID), m.input.Value())
			return nil, nil, m.input.Value()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd, nil
}

func (m *CreateChannelModal) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(cream).Render("Create Channel")
	idRow := lipgloss.JoinHorizontal(lipgloss.Top, "\nName: ", inputStyle.Render(m.input.View()))

	content := fmt.Sprintf("%s\n\n%s\n\n%s",
		title,
		idRow,
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

	protocol := "http://"
	if app.IsServerSecure {
		protocol = "https://"
	}

	token, err := app.LoadToken()
	if err != nil {

	}

	err = app.POST(
		payload,
		token,
		protocol+app.GetCurrentServerAddress()+app.JoinServerEndpoint,
		&response,
	)

}

func CreateChannel(serverID string, channelName string) {

	var response any

	payload := map[string]string{
		"serverID": serverID,
		"name":     channelName,
	}

	protocol := "http://"
	if app.IsServerSecure {
		protocol = "https://"
	}

	token, err := app.LoadToken()
	if err != nil {

	}

	err = app.POST(
		payload,
		token,
		protocol+app.GetCurrentServerAddress()+app.CreateChannelEndpoint,
		&response,
	)

}
