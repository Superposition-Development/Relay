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

	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#4a4a4a"))
	modalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cream).
			Background(lipgloss.Color("#121212")).
			Padding(1, 2)
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
	modalType            int
	activeModal          Modal
	scrollOffset         int
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

const (
	joinServerModal = iota
	createChannelModal
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
		activeModal:        nil,
	}
}

func (m *ChatScreen) Init() tea.Cmd {
	return tea.Batch(app.ListenForWSMsg(), tickCmd())
}

func (m *ChatScreen) Update(msg tea.Msg) (app.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case app.WebsocketMsg:
		switch msg.Type {
		case "recieveMessage":
			if m.activeServerIndex < 0 || m.activeChannelIndex < 0 {
				return m, nil
			}

			activeServerID := fmt.Sprintf("%v", app.ServerListToDataMap[m.activeServerIndex].ID)
			activeChannelID := fmt.Sprintf("%v", app.ChannelListToDataMap[m.activeChannelIndex].ID)

			serverID := fmt.Sprintf("%v", msg.Data["serverID"])
			channelID := fmt.Sprintf("%v", msg.Data["channelID"])

			var msgID int64
			if idFloat, ok := msg.Data["id"].(float64); ok {
				msgID = int64(idFloat)
			}

			var timestamp int64
			if tsFloat, ok := msg.Data["timestamp"].(float64); ok {
				timestamp = int64(tsFloat)
			}

			newMessage := app.Message{
				ID:        msgID,
				Username:  msg.Data["name"].(string),
				Content:   msg.Data["content"].(string),
				Timestamp: timestamp,
			}

			if serverID == activeServerID && channelID == activeChannelID {
				app.Messages = append(app.Messages, newMessage)
			}

			return m, app.ListenForWSMsg()
		case "newServer":
			app.Servers = GetServers()
			return m, app.ListenForWSMsg()
		case "newChannel":
			//eventually this needs to be changed to be the serverID associated with new channel
			app.Channels = GetChannels(app.CurrentServerID)
			return m, app.ListenForWSMsg()
		}

	case tickMsg:
		m.cursorBlink = !m.cursorBlink
		return m, tickCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		m.cursorBlink = true
		key := msg.String()

		if m.activeModal != nil {
			switch key {
			case "esc":
				m.activeModal = nil
				return m, nil
			}
			// fmt.Println("help!!")
			nextModal, _, result := m.activeModal.Update(msg)
			m.activeModal = nextModal

			if result != nil {
				switch v := result.(type) {
				case string:

					_ = v
				}
			}
			return m, nil
		}

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
				if m.selectedServerIndex <= 2 {
					switch m.selectedServerIndex {

					case 0:
						return m, func() tea.Msg {
							return app.ChangeScreenMsg{
								Screen: NewCreateServerScreen(m.height, m.width),
								Width:  m.width,
								Height: m.height,
							}
						}
					case 1:
						m.activeModal = NewJoinServerModal()
						m.modalType = joinServerModal

						return m, nil
					case 2:

						return m, nil
					}
				}

				m.activeServerIndex = m.selectedServerIndex
				app.CurrentServerID = app.ServerListToDataMap[m.activeServerIndex].ID
				app.Channels = GetChannels(app.CurrentServerID)
				m.focusedPanel = channelMenu
				app.Messages = nil
				break
			}

			if m.focusedPanel == channelMenu && len(app.Channels) > 0 {
				if m.selectedChannelIndex == 0 {
					m.activeModal = NewCreateChannelModal()
					m.modalType = joinServerModal

					return m, nil
				}
				m.activeChannelIndex = m.selectedChannelIndex
				m.focusedPanel = typingField
				m.inMenu = false
				serverID := app.CurrentServerID
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
			if m.focusedPanel == messages {
				if m.scrollOffset > 0 {
					m.scrollOffset--
				}
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
			if m.focusedPanel == messages {
				m.scrollOffset++
			}

		default:
			if (len(key) == 1 || strings.HasPrefix(key, " ")) && m.activeModal == nil {
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

	serverIDLabel := ""
	if m.activeServerIndex > 0 {
		serverIDLabel = " | ServerID: " + fmt.Sprintf("%v", app.CurrentServerID)
	}

	addr := app.GetCurrentServerAddress() + serverIDLabel
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
		m.scrollOffset,
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
	baseView := strings.Join(formattedContent, "\n")

	if m.activeModal != nil {

		return overlayModal(baseView, m.activeModal.View(), m.width, m.height)
	}

	return baseView
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

func chatBox(width, height, topLine int, title string, messages []app.Message, input string, cursorPos int, cursorBlink, isFocused bool, scrollOffset int) string {
	width, height = clamp(width, 2, width), clamp(height, 4, height)
	innerWidth := width - 2

	empty, horizontal := "│"+strings.Repeat(" ", innerWidth)+"│", "├"+strings.Repeat("─", innerWidth)+"┤"
	var formattedInputLines []string
	runes := []rune(input)
	currentLine, currentLineWidth := "┃ ", 2

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
			currentLine, currentLineWidth = "┃ "+renderedChar, 2+lipgloss.Width(charStr)
		} else {
			currentLine += renderedChar
			currentLineWidth += lipgloss.Width(charStr)
		}
		if r == '\n' {
			formattedInputLines = append(formattedInputLines, currentLine)
			currentLine, currentLineWidth = "┃ ", 2
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

	var allMessageLines []string
	senderStyle, timeStyle := lipgloss.NewStyle().Foreground(cream).Bold(true), lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))

	for _, msg := range messages {
		allMessageLines = append(allMessageLines, senderStyle.Render(msg.Username)+" "+timeStyle.Render(formatMessageTime(msg.Timestamp))+":")
		line := "  "
		for _, word := range strings.Fields(msg.Content) {
			if lipgloss.Width(line)+lipgloss.Width(word)+1 > innerWidth {
				allMessageLines = append(allMessageLines, line)
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
			allMessageLines = append(allMessageLines, line)
		}
	}

	totalLines := len(allMessageLines)
	maxOffset := clamp(totalLines-messageAreaHeight, 0, totalLines)
	scrollOffset = clamp(scrollOffset, 0, maxOffset)

	endIdx := totalLines - scrollOffset
	startIdx := endIdx - messageAreaHeight

	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx < 0 {
		endIdx = 0
	}

	visibleLines := allMessageLines[startIdx:endIdx]

	scrollbarThumbRow := -1
	showScrollbar := totalLines > messageAreaHeight

	if showScrollbar {
		scrollRatio := float64(scrollOffset) / float64(maxOffset)
		scrollbarThumbRow = int(float64(messageAreaHeight-1) * (1.0 - scrollRatio))
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

	thumbStyle := lipgloss.NewStyle().Foreground(cream)
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3a3a3a"))

	for i := 0; i < messageAreaHeight; i++ {
		row := (topLine + 1) + i
		if row >= dividerLine {
			break
		}

		msgContent := ""
		if i < len(visibleLines) {
			msgContent = visibleLines[i]
		}

		padLen := clamp(innerWidth-lipgloss.Width(msgContent), 0, innerWidth)
		rightChar := "│"

		if showScrollbar {
			if i == scrollbarThumbRow {
				rightChar = thumbStyle.Render("█")
			} else {
				rightChar = trackStyle.Render("│")
			}
		}

		lines[row] = "│" + msgContent + strings.Repeat(" ", padLen) + rightChar
	}

	lines[dividerLine] = horizontal

	inputStartLine := dividerLine + 1
	for i, inputLine := range formattedInputLines {
		row := inputStartLine + i
		if row >= height-1 {
			break
		}
		lines[row] = inputLine + strings.Repeat(" ", clamp(innerWidth-lipgloss.Width(inputLine), 0, innerWidth)) + " │"
	}

	return strings.Join(lines, "\n")
}

func SendMessage(m *ChatScreen) {
	token, err := app.LoadToken()
	if err != nil {

	}
	if m.inputBuffer != "" {
		payload := map[string]any{
			"serverID":  app.CurrentServerID,
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
		Name: "+ Create",
	}

	joinServerItem := app.Server{
		ID:   "Six Seven",
		Name: "+ Join",
	}

	selectDMItem := app.Server{
		ID:   "Six Seven",
		Name: "+ DM",
	}

	servers := append([]app.Server{createServerItem, joinServerItem, selectDMItem}, realServers...)

	app.ServerListToDataMap = make(map[int]app.Server, len(servers))
	for i, s := range servers {
		app.ServerListToDataMap[i] = s
	}

	return servers
}

func GetChannels(serverID any) []app.Channel {
	JWTCookie, err := app.LoadToken()
	if err != nil {
		fmt.Print(err)
	}
	reqPayload := GetChannelsRequest{
		ServerID: fmt.Sprintf("%v", serverID),
	}
	url := "http://" + app.GetCurrentServerAddress() + app.GetChannelEndpoint

	var realChannels []app.Channel
	if err := app.POST(reqPayload, JWTCookie, url, &realChannels); err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	createChannelItem := app.Channel{
		ID:   "Six Seven",
		Name: "+ Create",
	}

	channels := append([]app.Channel{createChannelItem}, realChannels...)

	app.ChannelListToDataMap = make(map[int]app.Channel, len(channels))
	for i, s := range channels {
		app.ChannelListToDataMap[i] = s
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
