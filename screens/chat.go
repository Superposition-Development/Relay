package screens

import (
	"strings"

	"Relay/app"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cream       = lipgloss.Color("#F5B978")
	borderStyle = lipgloss.NewStyle().Foreground(cream)
)

type ChatScreen struct {
	width        int
	height       int
	messageInput textarea.Model
}

func CreateChatScreen(h, w int) ChatScreen {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.MaxWidth = 20

	ta.Cursor.Style = lipgloss.NewStyle().Foreground(cream)
	ta.Cursor.TextStyle = lipgloss.NewStyle().Foreground(cream)

	return ChatScreen{
		height:       h,
		width:        w - 1,
		messageInput: ta,
	}
}

func (m ChatScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m ChatScreen) Update(msg tea.Msg) (app.Screen, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "ctrl+w":
			if app.CurrentInteractionMode == app.Write {
				app.CurrentInteractionMode = app.Default
			} else {
				app.CurrentInteractionMode = app.Write
			}

		case "ctrl+s":
			if app.CurrentInteractionMode == app.Write {
				SendMessage(&m)
			}

		case "enter":
			if app.CurrentInteractionMode != app.Write {
				SendMessage(&m)
				return m, cmd
			}

		case "tab":
			if app.CurrentInteractionMode == app.Write {
				value := m.messageInput.Value()
				m.messageInput.SetValue(value + "    ")
			}
		}
	}

	m.messageInput, cmd = m.messageInput.Update(msg)

	lines := strings.Count(m.messageInput.Value(), "\n") + 1
	m.messageInput.SetHeight(clamp(lines, 1, 5))

	return m, cmd
}

func titledBox(title string, width, height int) string {
	minWidth := len(title) + 5
	width = clamp(width, minWidth, width)
	height = clamp(height, 2, height)

	top := "┌─ " + title + " " + strings.Repeat("─", width-len(title)-5) + "┐"
	empty := "│" + strings.Repeat(" ", width-2) + "│"
	bottom := "└" + strings.Repeat("─", width-2) + "┘"

	lines := []string{top}
	for i := 0; i < height-2; i++ {
		lines = append(lines, empty)
	}
	lines = append(lines, bottom)

	return strings.Join(lines, "\n")
}

func horizontalLine(width int) string {
	if width < 1 {
		return ""
	}
	return strings.Repeat("─", width)
}

func chatBox(width, height, topLine int, text, input string) string {
	width = clamp(width, 2, width)
	height = clamp(height, 4, height)

	innerWidth := width - 2
	empty := "│" + strings.Repeat(" ", innerWidth) + "│"
	horizontal := "├" + horizontalLine(innerWidth) + "┤"

	lines := make([]string, height)
	for i := 0; i < height; i++ {
		lines[i] = empty
	}

	if topLine >= 0 && topLine < height {
		lines[topLine] = horizontal
	}

	if topLine > 0 && text != "" {
		textRow := []rune(empty)
		textRunes := []rune(text)
		startX := 1

		for i, r := range textRunes {
			x := startX + i
			if x >= innerWidth+1 {
				break
			}
			textRow[x] = r
		}
		lines[topLine-1] = string(textRow)
	}

	inputLines := strings.Split(input, "\n")
	inputHeight := len(inputLines)

	dividerLine := height - inputHeight - 2
	if dividerLine <= topLine {
		dividerLine = topLine + 1
	}

	lines[dividerLine] = horizontal
	inputStartLine := dividerLine + 1

	for i, inputLine := range inputLines {
		row := inputStartLine + i
		if row >= height-1 {
			break
		}

		if lipgloss.Width(inputLine) > innerWidth {
			inputLine = inputLine[:innerWidth]
		}

		padding := clamp(innerWidth-lipgloss.Width(inputLine), 0, innerWidth)

		lines[row] = "│" + inputLine + strings.Repeat(" ", padding) + "│"
	}

	return strings.Join(lines, "\n")
}

func (m ChatScreen) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	serversWidth := 12
	channelsWidth := 20
	rightPadding := 10
	panelHeight := clamp(m.height-5, 4, m.height)

	title := "Relay"
	headerLineWidth := clamp(m.width-len(title)-5, 0, m.width)

	top := "┌─ " + title + " " + horizontalLine(headerLineWidth) + "┐"

	string1 := app.GetCurrentServerAddress()
	string2 := "string 2"

	spacing := clamp(m.width-len(string1)-len(string2)-2, 1, m.width)
	middle := "│" + string1 + strings.Repeat(" ", spacing) + string2 + "│"

	leftDividerX := serversWidth + channelsWidth
	rightDividerX := m.width - rightPadding - 3

	servers := titledBox("Servers", serversWidth, panelHeight)
	channels := titledBox("Channels", channelsWidth, panelHeight)

	chatWidth := clamp(rightDividerX-leftDividerX, 2, rightDividerX)
	chat := chatBox(chatWidth, panelHeight, 1, "channel name", m.messageInput.Value())

	leftPanels := lipgloss.JoinHorizontal(lipgloss.Top, servers, channels)
	fullPanels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanels, chat)

	separatorWidth := m.width - 2
	separatorChars := []rune(strings.Repeat("─", separatorWidth))

	leftSeparatorX := leftDividerX - 1
	if leftSeparatorX >= 0 && leftSeparatorX < len(separatorChars) {
		separatorChars[leftSeparatorX+1] = '┬'
	}

	rightSeparatorX := rightDividerX - 1
	if rightSeparatorX >= 0 && rightSeparatorX < len(separatorChars) {
		separatorChars[rightSeparatorX] = '┬'
	}

	separator := "├" + string(separatorChars) + "┤"

	content := []string{top, middle, separator}
	panelLines := strings.Split(fullPanels, "\n")

	for _, line := range panelLines {
		padding := clamp(m.width-lipgloss.Width(line)-2, 0, m.width)
		content = append(content, "│"+line+strings.Repeat(" ", padding)+"│")
	}

	bottomWidth := m.width - 2
	bottomChars := []rune(strings.Repeat("─", bottomWidth))

	leftBottomX := leftDividerX - 1
	if leftBottomX >= 0 && leftBottomX < len(bottomChars) {
		bottomChars[leftBottomX+1] = '┴'
	}

	rightBottomX := rightDividerX - 1
	if rightBottomX >= 0 && rightBottomX < len(bottomChars) {
		bottomChars[rightBottomX] = '┴'
	}

	bottom := "└" + string(bottomChars) + "┘"
	content = append(content, bottom)

	return borderStyle.Render(strings.Join(content, "\n"))
}

func SendMessage(m *ChatScreen) {
	if m.messageInput.Value() != "" {
		m.messageInput.Reset()
	}
}

func clamp(val, minVal, maxVal int) int {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}
