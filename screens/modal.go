package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var inputStyle = lipgloss.NewStyle().
	Width(40).
	Height(1).
	Border(lipgloss.NormalBorder()).
	BorderForeground(cream).
	Foreground(cream)

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

func (m ChatScreen) renderJoinServerModal() string {
	label := "\nServerID: "
	serverIDLabel := textinput.New()
	serverIDLabel.Prompt = ""
	title := lipgloss.NewStyle().Bold(true).Foreground(cream).Render("Create Server")
	idRow := lipgloss.JoinHorizontal(lipgloss.Top, label, inputStyle.Render(serverIDLabel.View()))

	content := fmt.Sprintf(
		"%s\n\n"+
			idRow+
			"%s",
		title,
		"\n"+lipgloss.NewStyle().Foreground(lipgloss.Color("#767676")).Render("Press Esc or 'q' to close"),
	)

	return modalBoxStyle.Render(content)
}

func (m ChatScreen) renderCreateChannelModal() string {
	label := "\nChannel Name: "
	serverIDLabel := textinput.New()
	serverIDLabel.Prompt = ""
	title := lipgloss.NewStyle().Bold(true).Foreground(cream).Render("Create Channel")
	idRow := lipgloss.JoinHorizontal(lipgloss.Top, label, inputStyle.Render(serverIDLabel.View()))

	content := fmt.Sprintf(
		"%s\n\n"+
			idRow+
			"%s",
		title,
		"\n"+lipgloss.NewStyle().Foreground(lipgloss.Color("#767676")).Render("Press Esc or 'q' to close"),
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
