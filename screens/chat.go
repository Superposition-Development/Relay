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
	width                int
	height               int
	inputBuffer          string
	cursorPos            int
	cursorBlink          bool
	selectedChannelIndex int
	activeChannelIndex   int
	focusedPanel         int
	inMenu               bool
	selectedServerIndex  int
	activeServerIndex    int
}

type GetChannelsRequest struct {
	ServerID any `json:"serverID"`
}

type GetMessagesRequest struct {
	ServerID  any `json:"serverID"`
	ChannelID any `json:"channelID"`
	MessageID any `json:"messageID"`
	Ascending any `json:"ascending"`
	MoreThan  any `json:"moreThan"`
}

const (
	profileMenu = iota
	serverMenu
	channelMenu
	messages
	typingField
)

func CreateChatScreen(h, w int) *ChatScreen {
	app.Servers = GetServers()
	return &ChatScreen{
		height:             h,
		width:              w - 1,
		cursorBlink:        true,
		focusedPanel:       profileMenu,
		activeChannelIndex: -1,
		activeServerIndex:  -1,
	}
}

func (m *ChatScreen) Init() tea.Cmd {
	return tea.Batch(app.ListenForWSMsg(), tickCmd())
}

func (m *ChatScreen) Update(msg tea.Msg) (app.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case app.WebsocketMsg:
		if m.activeServerIndex < 0 || m.activeChannelIndex < 0 {
			return m, nil
		}
		activeServerID := fmt.Sprintf("%v", app.ServerListToDataMap[m.activeServerIndex].ID)
		activeChannelID := fmt.Sprintf("%v", app.ChannelListToDataMap[m.activeChannelIndex].ID)

		if msg.ServerID == activeServerID && msg.ChannelID == activeChannelID {
			app.Messages = append(app.Messages, msg.Message)
		}

		return m, app.ListenForWSMsg()

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
		case "esc":
			m.inMenu = false
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
				SendMessage(m)
			}

		case "enter":
			if m.focusedPanel == typingField {
				if app.CurrentInteractionMode == app.Write {
					m.insertRune('\n')
				} else {
					SendMessage(m)
					return m, nil
				}
			}

			if !m.inMenu {
				if m.focusedPanel == serverMenu || m.focusedPanel == channelMenu {
					m.inMenu = true
				}
				break
			}

			if m.focusedPanel == serverMenu && len(app.Servers) > 0 {
				if m.selectedServerIndex == 0 {
					return m, func() tea.Msg {
						return app.ChangeScreenMsg{
							Screen: NewCreateServerScreen(m.height, m.width),
							Width:  m.width,
							Height: m.height,
						}
					}
				}

				m.activeServerIndex = m.selectedServerIndex
				app.Channels = GetChannels(app.ServerListToDataMap[m.activeServerIndex].ID)
				m.focusedPanel = channelMenu
				app.Messages = nil
				break
			}

			if m.focusedPanel == channelMenu && len(app.Channels) > 0 {
				m.activeChannelIndex = m.selectedChannelIndex
				m.focusedPanel = typingField
				m.inMenu = false
				serverID := app.ServerListToDataMap[m.activeServerIndex].ID
				channelID := app.ChannelListToDataMap[m.activeChannelIndex].ID

				app.Messages = app.ReverseMessages(GetMessages(
					serverID,
					channelID,
					"0", false, true,
				))
				break
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
			if m.inMenu && m.focusedPanel == serverMenu && len(app.Servers) > 0 {
				m.selectedServerIndex = (m.selectedServerIndex + 1) % len(app.Servers)
			}
			if m.inMenu && m.focusedPanel == channelMenu && len(app.Channels) > 0 {
				m.selectedChannelIndex = (m.selectedChannelIndex + 1) % len(app.Channels)
			}

		case "up":
			if m.inMenu && m.focusedPanel == serverMenu && len(app.Servers) > 0 {
				m.selectedServerIndex--
				if m.selectedServerIndex < 0 {
					m.selectedServerIndex = len(app.Servers) - 1
				}
			}
			if m.inMenu && m.focusedPanel == channelMenu && len(app.Channels) > 0 {
				m.selectedChannelIndex--
				if m.selectedChannelIndex < 0 {
					m.selectedChannelIndex = len(app.Channels) - 1
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

	servers := renderListBox("Servers", app.Servers, func(s app.Server) string { return s.Name },
		serversWidth, panelHeight, m.focusedPanel == serverMenu, m.inMenu, m.selectedServerIndex, m.activeServerIndex, m.cursorBlink)

	channels := renderListBox("Channels", app.Channels, func(c app.Channel) string { return "# " + c.Name },
		channelsWidth, panelHeight, m.focusedPanel == channelMenu, m.inMenu, m.selectedChannelIndex, m.activeChannelIndex, m.cursorBlink)

	chat := chatBox(
		chatWidth,
		panelHeight,
		1,
		app.ChannelListToDataMap[m.activeChannelIndex].Name,
		app.Messages,
		m.inputBuffer,
		m.cursorPos,
		m.cursorBlink,
		m.focusedPanel == typingField,
	)
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

func formatMessageTime(epochSecs int64) string {
	return time.Unix(epochSecs, 0).Local().Format("3:04 PM (01/02/2006)")
}

func chatBox(width, height, topLine int, title string, messages []app.Message, input string, cursorPos int, cursorBlink bool, isFocused bool) string {
	width = clamp(width, 2, width)
	height = clamp(height, 4, height)
	innerWidth := width - 2

	empty := "│" + strings.Repeat(" ", innerWidth) + "│"
	horizontal := "├" + strings.Repeat("─", innerWidth) + "┤"

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
		if cursorBlink && isFocused {
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

	dividerLine := clamp(height-len(formattedInputLines)-2, topLine+1, height-1)
	messageAreaHeight := dividerLine - (topLine + 1)

	var messageLines []string
	senderStyle := lipgloss.NewStyle().Foreground(cream).Bold(true)
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))

	for _, msg := range messages {
		timeStr := formatMessageTime(msg.Timestamp)

		messageLines = append(messageLines, senderStyle.Render(msg.Username)+" "+timeStyle.Render(timeStr)+":")

		words := strings.Fields(msg.Content)
		line := "  "
		for _, word := range words {
			if lipgloss.Width(line)+lipgloss.Width(word)+1 > innerWidth {
				messageLines = append(messageLines, line)
				line = "  " + word
			} else {
				if line == "  " {
					line += word
				} else {
					line += " " + word
				}
			}
		}
		if strings.TrimSpace(line) != "" {
			messageLines = append(messageLines, line)
		}
	}

	if len(messageLines) > messageAreaHeight {
		messageLines = messageLines[len(messageLines)-messageAreaHeight:]
	}

	lines := make([]string, height)
	for i := range lines {
		lines[i] = empty
	}

	if topLine > 0 && title != "" {
		textRow := []rune(empty)
		for i, r := range []rune(title) {
			if i+1 >= innerWidth+1 {
				break
			}
			textRow[i+1] = r
		}
		lines[topLine-1] = string(textRow)
	}

	if topLine >= 0 && topLine < height {
		lines[topLine] = horizontal
	}

	for i, msgLine := range messageLines {
		row := (topLine + 1) + i
		if row >= dividerLine {
			break
		}
		pad := clamp(innerWidth-lipgloss.Width(msgLine), 0, innerWidth)
		lines[row] = "│" + msgLine + strings.Repeat(" ", pad) + "│"
	}

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
	token, err := app.LoadToken()
	if err != nil {

	}
	if m.inputBuffer != "" {
		payload := map[string]any{
			"serverID":  app.ServerListToDataMap[m.activeServerIndex].ID,
			"channelID": fmt.Sprintf("%v", app.ChannelListToDataMap[m.activeChannelIndex].ID),
			"content":   m.inputBuffer,
			"authKey":   token,
			"message":   "sendMessage",
		}

		app.SendWebsocketJSON(payload)
		m.inputBuffer = ""
		m.cursorPos = 0
	}
}

func GetServers() []app.Server {
	JWTCookie, err := app.LoadToken()
	if err != nil {
		fmt.Print(err)
	}
	res := app.GET(JWTCookie, "http://"+app.GetCurrentServerAddress()+app.GetServerEndpoint)
	if res.Error != nil {
		fmt.Println("Error:", res.Error)
		return nil
	}

	var realServers []app.Server
	if err := json.Unmarshal(res.Data, &realServers); err != nil {
		fmt.Println("Error decoding servers:", err)
		return nil
	}

	createServerItem := app.Server{
		ID:   "Six Seven",
		Name: "+ Create Server",
	}

	servers := append([]app.Server{createServerItem}, realServers...)

	app.ServerListToDataMap = make(map[int]app.Server, len(servers))
	for i, s := range servers {
		app.ServerListToDataMap[i] = s
	}

	return servers
}

func GetChannels(serverID any) []app.Channel {
	JWTCookie, err := app.LoadToken()
	if err != nil || JWTCookie == "" {
		return nil
	}

	reqPayload := GetChannelsRequest{
		ServerID: fmt.Sprintf("%v", serverID),
	}
	url := "http://" + app.GetCurrentServerAddress() + app.GetChannelEndpoint

	var channels []app.Channel
	if err := app.POST(reqPayload, JWTCookie, url, &channels); err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	app.ChannelListToDataMap = make(map[int]app.Channel, len(channels))
	for i, c := range channels {

		app.ChannelListToDataMap[i] = c
	}

	return channels
}

func GetMessages(serverID any, channelID any, messageID any, ascending any, moreThan any) []app.Message {

	JWTCookie, err := app.LoadToken()
	if err != nil || JWTCookie == "" {
		return nil
	}

	reqPayload := GetMessagesRequest{
		ServerID:  fmt.Sprintf("%v", serverID),
		ChannelID: fmt.Sprintf("%v", channelID),
		MessageID: fmt.Sprintf("%v", messageID),
		Ascending: fmt.Sprintf("%v", ascending),
		MoreThan:  fmt.Sprintf("%v", moreThan),
	}
	url := "http://" + app.GetCurrentServerAddress() + app.GetMessagesEndpoint

	var messages []app.Message
	if err := app.POST(reqPayload, JWTCookie, url, &messages); err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return messages
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
