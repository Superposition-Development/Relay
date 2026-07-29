package screens

import (
	"Relay/app"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cream       = lipgloss.Color("#F5B978")
	borderStyle = lipgloss.NewStyle().Foreground(cream)
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a1a")).Background(cream)
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type ChatScreen struct {
	width       int
	height      int
	inputBuffer string
	cursorPos   int
	cursorBlink bool
}

func CreateChatScreen(h, w int) ChatScreen {
	return ChatScreen{
		height:      h,
		width:       w - 1,
		inputBuffer: "",
		cursorPos:   0,
		cursorBlink: true,
	}
}

func (m ChatScreen) Init() tea.Cmd {
	return tickCmd()
}

func (m ChatScreen) Update(msg tea.Msg) (app.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.cursorBlink = !m.cursorBlink
		return m, tickCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		GetServers()
		m.cursorBlink = true

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
			if app.CurrentInteractionMode == app.Write {
				runes := []rune(m.inputBuffer)
				m.inputBuffer = string(runes[:m.cursorPos]) + "\n" + string(runes[m.cursorPos:])
				m.cursorPos++
			} else {
				SendMessage(&m)
				return m, nil
			}

		case "tab":
			if app.CurrentInteractionMode == app.Write {
				runes := []rune(m.inputBuffer)
				m.inputBuffer = string(runes[:m.cursorPos]) + "    " + string(runes[m.cursorPos:])
				m.cursorPos += 4
			}

		case "backspace":
			if m.cursorPos > 0 {
				runes := []rune(m.inputBuffer)
				m.inputBuffer = string(runes[:m.cursorPos-1]) + string(runes[m.cursorPos:])
				m.cursorPos--
			}

		case "left":
			if m.cursorPos > 0 {
				m.cursorPos--
			}

		case "right":
			if m.cursorPos < len([]rune(m.inputBuffer)) {
				m.cursorPos++
			}

		default:
			s := msg.String()
			if len(s) == 1 || strings.HasPrefix(s, " ") {
				runes := []rune(m.inputBuffer)
				m.inputBuffer = string(runes[:m.cursorPos]) + s + string(runes[m.cursorPos:])
				m.cursorPos += len([]rune(s))
			}
		}
	}

	return m, nil
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

func chatBox(width, height, topLine int, text, input string, cursorPos int, cursorBlink bool) string {
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
	creamPrefix := borderStyle.Render("")

	var formattedInputLines []string
	runes := []rune(input)

	currentLine := "┃ "
	currentLineWidth := 2

	for i, r := range runes {
		charStr := string(r)
		if r == '\n' {
			charStr = " "
		}

		renderedChar := charStr
		if i == cursorPos && cursorBlink {
			renderedChar = cursorStyle.Render(charStr) + creamPrefix
		}

		if currentLineWidth+lipgloss.Width(charStr) > innerWidth {
			formattedInputLines = append(formattedInputLines, currentLine)
			currentLine = "┃ " + renderedChar
			currentLineWidth = 2 + lipgloss.Width(charStr)
		} else {
			currentLine += renderedChar
			currentLineWidth += lipgloss.Width(charStr)
		}

		if r == '\n' {
			formattedInputLines = append(formattedInputLines, currentLine)
			currentLine = "┃ "
			currentLineWidth = 2
		}
	}

	if cursorPos >= len(runes) {
		cursorChar := " "
		if cursorBlink {
			cursorChar = cursorStyle.Render(" ") + creamPrefix
		}

		if currentLineWidth+1 > innerWidth {
			formattedInputLines = append(formattedInputLines, currentLine)
			currentLine = "┃ " + cursorChar
		} else {
			currentLine += cursorChar
		}
	}

	formattedInputLines = append(formattedInputLines, currentLine)

	inputHeight := len(formattedInputLines)
	dividerLine := height - inputHeight - 2
	if dividerLine <= topLine {
		dividerLine = topLine + 1
	}

	lines[dividerLine] = horizontal
	inputStartLine := dividerLine + 1

	for i, inputLine := range formattedInputLines {
		row := inputStartLine + i
		if row >= height-1 {
			break
		}

		visibleWidth := lipgloss.Width(inputLine)
		padding := clamp(innerWidth-visibleWidth, 0, innerWidth)
		lines[row] = inputLine + strings.Repeat(" ", padding) + " │"
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
	chat := chatBox(chatWidth, panelHeight, 1, "channel name", m.inputBuffer, m.cursorPos, m.cursorBlink)

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
		visibleLineWidth := lipgloss.Width(line)
		padding := clamp(m.width-visibleLineWidth-2, 0, m.width)
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
	if m.inputBuffer != "" {
		m.inputBuffer = ""
		m.cursorPos = 0
	}
}

type Server struct {
	ID   any    `json:"id"`
	PFP  string `json:"pfp"`
	Name string `json:"name"`
}

func GetServers() {
	JWTCookie, err := app.LoadToken()
	if err != nil {
		fmt.Print(err)
	}
	res := app.GET(JWTCookie, "http://"+app.GetCurrentServerAddress()+app.GetServerEndpoint)
	if res.Error != nil {
		fmt.Println("Error:", res.Error)
		return
	}

	var servers []Server
	if err := json.Unmarshal(res.Data, &servers); err != nil {
		fmt.Println("Error decoding servers:", err)
		return
	}

	for _, s := range servers {
		fmt.Printf("ID: %s | Name: %s | PFP: %s\n", s.ID, s.Name, s.PFP)
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
