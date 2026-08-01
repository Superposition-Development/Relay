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
	cream               = lipgloss.Color("#F5B978")
	borderStyle         = lipgloss.NewStyle().Foreground(cream)
	cursorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a1a")).Background(cream)
	selectedBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffdea5"))
	activeStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a1a")).Background(lipgloss.Color("#ffdea5")).Bold(true)
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type ChatScreen struct {
	width               int
	height              int
	inputBuffer         string
	cursorPos           int
	cursorBlink         bool
	servers             []Server
	channels            []Channel
	selectedChannelIdx  int
	activeChannelIdx    int
	focusedPanel        int
	inMenu              bool
	selectedServerIndex int
	activeServerIndex   int
}

type Channel struct {
	ID   any    `json:"id"`
	Name string `json:"name"`
}

type GetChannelsRequest struct {
	ServerID any `json:"serverID"`
}

type Server struct {
	ID   any    `json:"id"`
	PFP  string `json:"pfp"`
	Name string `json:"name"`
}

const (
	profileMenu = iota
	serverMenu
	channelMenu
	messages
	typingField
)

func CreateChatScreen(h, w int) ChatScreen {
	return ChatScreen{
		height:       h,
		width:        w - 1,
		cursorBlink:  true,
		servers:      GetServers(),
		focusedPanel: profileMenu,
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
		m.cursorBlink = true
		key := msg.String()

		switch key {
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
			if !m.inMenu {
				if m.focusedPanel == serverMenu || m.focusedPanel == channelMenu {
					m.inMenu = true
				}
				break
			}

			if m.focusedPanel == serverMenu && len(m.servers) > 0 {
				m.activeServerIndex = m.selectedServerIndex
				m.channels = GetChannels(app.ServerListToIDMap[m.activeServerIndex])
				m.inMenu = false
			}

			if m.focusedPanel == typingField {
				if app.CurrentInteractionMode == app.Write {
					m.insertRune('\n')
				} else {
					SendMessage(&m)
					return m, nil
				}
			}

		case "tab":
			if app.CurrentInteractionMode == app.Write {
				m.insertString("    ")
				m.focusedPanel = typingField
				break
			}
			if !m.inMenu {
				m.focusedPanel = (m.focusedPanel + 1) % (typingField + 1)
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

		case "down":
			if m.inMenu && m.focusedPanel == serverMenu && len(m.servers) > 0 {
				m.selectedServerIndex = (m.selectedServerIndex + 1) % len(m.servers)
			}

		case "up":
			if m.inMenu && m.focusedPanel == serverMenu && len(m.servers) > 0 {
				m.selectedServerIndex--
				if m.selectedServerIndex < 0 {
					m.selectedServerIndex = len(m.servers) - 1
				}
			}

		default:
			if len(key) == 1 || strings.HasPrefix(key, " ") {
				m.insertString(key)
			}
		}
	}

	return m, nil
}

func (m *ChatScreen) insertString(s string) {
	runes := []rune(m.inputBuffer)
	m.inputBuffer = string(runes[:m.cursorPos]) + s + string(runes[m.cursorPos:])
	m.cursorPos += len([]rune(s))
}

func (m *ChatScreen) insertRune(r rune) {
	m.insertString(string(r))
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
	headerWidth := clamp(m.width-len(title)-5, 0, m.width)
	top := borderStyle.Render("┌─ " + title + " " + strings.Repeat("─", headerWidth) + "┐")

	addr := app.GetCurrentServerAddress()
	userID := app.CurrentUserID
	spacing := clamp(m.width-len(addr)-len(userID)-2, 1, m.width)
	middle := borderStyle.Render("│" + addr + strings.Repeat(" ", spacing) + userID + "│")

	leftDividerX := serversWidth + channelsWidth
	rightDividerX := m.width - rightPadding - 3
	chatWidth := clamp(rightDividerX-leftDividerX, 2, rightDividerX)

	servers := renderListBox("Servers", m.servers, func(s Server) string { return s.Name },
		serversWidth, panelHeight, m.focusedPanel == serverMenu, m.inMenu, m.selectedServerIndex, m.activeServerIndex, m.cursorBlink)

	channels := renderListBox("Channels", m.channels, func(c Channel) string { return "# " + c.Name },
		channelsWidth, panelHeight, m.focusedPanel == channelMenu, m.inMenu, m.selectedChannelIdx, m.activeChannelIdx, m.cursorBlink)

	chat := chatBox(chatWidth, panelHeight, 1, "channel name", m.inputBuffer, m.cursorPos, m.cursorBlink)

	fullPanels := lipgloss.JoinHorizontal(lipgloss.Top, servers, channels, chat)

	sepRunes := []rune(strings.Repeat("─", m.width-2))
	if idx := leftDividerX - 1; idx >= 0 && idx < len(sepRunes) {
		sepRunes[idx+1] = '┬'
	}
	if idx := rightDividerX - 1; idx >= 0 && idx < len(sepRunes) {
		sepRunes[idx] = '┬'
	}
	separator := borderStyle.Render("├" + string(sepRunes) + "┤")

	botRunes := []rune(strings.Repeat("─", m.width-2))
	if idx := leftDividerX - 1; idx >= 0 && idx < len(botRunes) {
		botRunes[idx+1] = '┴'
	}
	if idx := rightDividerX - 1; idx >= 0 && idx < len(botRunes) {
		botRunes[idx] = '┴'
	}
	bottom := borderStyle.Render("└" + string(botRunes) + "┘")

	panelLines := strings.Split(fullPanels, "\n")
	var formattedContent []string
	formattedContent = append(formattedContent, top, middle, separator)

	for _, line := range panelLines {
		pad := clamp(m.width-lipgloss.Width(line)-2, 0, m.width)
		formattedContent = append(formattedContent, borderStyle.Render("│"+line+strings.Repeat(" ", pad)+"│"))
	}

	formattedContent = append(formattedContent, bottom)
	return strings.Join(formattedContent, "\n")
}

func renderListBox[T any](
	title string,
	items []T,
	getName func(T) string,
	width, height int,
	isFocused, inMenu bool,
	selectedIndex, activeIndex int,
	blink bool,
) string {
	width = clamp(width, len(title)+5, width)
	height = clamp(height, 2, height)
	innerWidth := width - 2

	style := borderStyle
	if isFocused {
		style = selectedBorderStyle
	}

	top := "┌─ " + title + " " + strings.Repeat("─", width-len(title)-5) + "┐"
	bottom := "└" + strings.Repeat("─", innerWidth) + "┘"

	lines := []string{top}

	for i := 0; i < height-2; i++ {
		var content string
		if i < len(items) {
			content = getName(items[i])
		}

		runes := []rune(content)
		if len(runes) > innerWidth-1 {
			runes = runes[:innerWidth-1]
		}
		content = string(runes)

		paddingLen := clamp(innerWidth-lipgloss.Width(content), 0, innerWidth)
		paddedContent := content + strings.Repeat(" ", paddingLen)

		if inMenu && isFocused && i == selectedIndex {
			if blink {
				paddedContent = cursorStyle.Render(paddedContent)
			} else {
				paddedContent = lipgloss.NewStyle().Foreground(cream).Render(paddedContent)
			}
		} else if i == activeIndex && len(items) > 0 {
			paddedContent = activeStyle.Render(paddedContent)
		}

		lines = append(lines, "│"+paddedContent+"│")
	}

	lines = append(lines, bottom)
	return style.Render(strings.Join(lines, "\n"))
}

func chatBox(width, height, topLine int, text, input string, cursorPos int, cursorBlink bool) string {
	width = clamp(width, 2, width)
	height = clamp(height, 4, height)
	innerWidth := width - 2

	empty := "│" + strings.Repeat(" ", innerWidth) + "│"
	horizontal := "├" + strings.Repeat("─", innerWidth) + "┤"

	lines := make([]string, height)
	for i := range lines {
		lines[i] = empty
	}

	if topLine >= 0 && topLine < height {
		lines[topLine] = horizontal
	}

	if topLine > 0 && text != "" {
		textRow := []rune(empty)
		for i, r := range []rune(text) {
			if i+1 >= innerWidth+1 {
				break
			}
			textRow[i+1] = r
		}
		lines[topLine-1] = string(textRow)
	}

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
			renderedChar = cursorStyle.Render(charStr) + borderStyle.Render("")
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
			cursorChar = cursorStyle.Render(" ") + borderStyle.Render("")
		}

		if currentLineWidth+1 > innerWidth {
			formattedInputLines = append(formattedInputLines, currentLine)
			currentLine = "┃ " + cursorChar
		} else {
			currentLine += cursorChar
		}
	}
	formattedInputLines = append(formattedInputLines, currentLine)

	dividerLine := clamp(height-len(formattedInputLines)-2, topLine+1, height)
	lines[dividerLine] = horizontal
	inputStartLine := dividerLine + 1

	for i, inputLine := range formattedInputLines {
		row := inputStartLine + i
		if row >= height-1 {
			break
		}
		padding := clamp(innerWidth-lipgloss.Width(inputLine), 0, innerWidth)
		lines[row] = inputLine + strings.Repeat(" ", padding) + " │"
	}

	return strings.Join(lines, "\n")
}

func SendMessage(m *ChatScreen) {
	if m.inputBuffer != "" {
		m.inputBuffer = ""
		m.cursorPos = 0
	}
}

func GetServers() []Server {
	JWTCookie, err := app.LoadToken()
	if err != nil {
		fmt.Print(err)
	}
	res := app.GET(JWTCookie, "http://"+app.GetCurrentServerAddress()+app.GetServerEndpoint)
	if res.Error != nil {
		fmt.Println("Error:", res.Error)
		return nil
	}

	var servers []Server
	if err := json.Unmarshal(res.Data, &servers); err != nil {
		fmt.Println("Error decoding servers:", err)
		return nil
	}

	app.ServerListToIDMap = make(map[int]any, len(servers))
	for i, s := range servers {
		app.ServerListToIDMap[i] = s.ID
	}
	return servers
}

func GetChannels(serverID any) []Channel {
	JWTCookie, err := app.LoadToken()
	if err != nil || JWTCookie == "" {
		return nil
	}

	reqPayload := GetChannelsRequest{
		ServerID: fmt.Sprintf("%v", serverID),
	}
	url := "http://" + app.GetCurrentServerAddress() + app.GetChannelEndpoint

	var channels []Channel
	if err := app.POST(reqPayload, JWTCookie, url, &channels); err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return channels
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
