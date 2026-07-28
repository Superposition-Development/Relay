package screens

import (
	"Relay/app"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cream = lipgloss.Color("#F5B978")

	borderStyle = lipgloss.NewStyle().
			Foreground(cream)
)

type ChatScreen struct {
	width  int
	height int
}

func CreateChatScreen(h, w int) ChatScreen {
	return ChatScreen{
		height: h,
		width:  w - 1,
	}
}

func (m ChatScreen) Init() tea.Cmd {
	return nil
}

func (m ChatScreen) Update(msg tea.Msg) (app.Screen, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func titledBox(title string, width, height int) string {

	minWidth := len(title) + 5

	if width < minWidth {
		width = minWidth
	}

	if height < 2 {
		height = 2
	}

	top :=
		"┌─ " +
			title +
			" " +
			strings.Repeat(
				"─",
				width-len(title)-5,
			) +
			"┐"

	empty :=
		"│" +
			strings.Repeat(
				" ",
				width-2,
			) +
			"│"

	bottom :=
		"└" +
			strings.Repeat(
				"─",
				width-2,
			) +
			"┘"

	lines := []string{top}

	for i := 0; i < height-2; i++ {
		lines = append(
			lines,
			empty,
		)
	}

	lines = append(
		lines,
		bottom,
	)

	return strings.Join(
		lines,
		"\n",
	)
}

func horizontalLine(width int) string {

	if width < 1 {
		return ""
	}

	return strings.Repeat(
		"─",
		width,
	)
}

func chatBox(
	width int,
	height int,
	topLine int,
	bottomLine int,
	text string,
) string {

	if width < 2 {
		width = 2
	}

	if height < 1 {
		height = 1
	}

	innerWidth := width - 2

	empty :=
		"│" +
			strings.Repeat(
				" ",
				innerWidth,
			) +
			"│"

	horizontal :=
		"├" +
			horizontalLine(
				innerWidth,
			) +
			"┤"

	lines := make([]string, height)

	for i := 0; i < height; i++ {

		if i == topLine {
			lines[i] = horizontal
			continue
		}

		if i == bottomLine {
			lines[i] = horizontal
			continue
		}

		lines[i] = empty
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

	return strings.Join(
		lines,
		"\n",
	)
}

func (m ChatScreen) View() string {

	if m.width <= 0 ||
		m.height <= 0 {

		return ""
	}

	serversWidth := 12

	channelsWidth := 20

	rightPadding := 10

	panelHeight :=
		m.height -
			5

	if panelHeight < 2 {
		panelHeight = 2
	}

	title := "Relay"

	headerLineWidth :=
		m.width -
			len(title) -
			5

	if headerLineWidth < 0 {
		headerLineWidth = 0
	}

	top :=
		"┌─ " +
			title +
			" " +
			horizontalLine(
				headerLineWidth,
			) +
			"┐"

	string1 := app.GetCurrentServerAddress()

	string2 := "string 2"

	spacing :=
		m.width -
			len(string1) -
			len(string2) -
			2

	if spacing < 1 {
		spacing = 1
	}

	middle :=
		"│" +
			string1 +
			strings.Repeat(
				" ",
				spacing,
			) +
			string2 +
			"│"

	leftDividerX :=
		serversWidth +
			channelsWidth

	rightDividerX :=
		m.width -
			rightPadding -
			3

	servers :=
		titledBox(
			"Servers",
			serversWidth,
			panelHeight,
		)

	channels :=
		titledBox(
			"Channels",
			channelsWidth,
			panelHeight,
		)

	chatWidth :=
		rightDividerX -
			leftDividerX

	if chatWidth < 2 {
		chatWidth = 2
	}

	chat :=
		chatBox(
			chatWidth,
			panelHeight,
			1,
			panelHeight-2,
			"channel name",
		)

	leftPanels :=
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			servers,
			channels,
		)

	fullPanels :=
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftPanels,
			chat,
		)

	separatorWidth :=
		m.width -
			2

	separatorChars :=
		[]rune(
			strings.Repeat(
				"─",
				separatorWidth,
			),
		)

	leftSeparatorX :=
		leftDividerX -
			1

	if leftSeparatorX >= 0 &&
		leftSeparatorX < len(separatorChars) {

		separatorChars[leftSeparatorX+1] = '┬'
	}

	rightSeparatorX :=
		rightDividerX -
			1

	if rightSeparatorX >= 0 &&
		rightSeparatorX < len(separatorChars) {

		separatorChars[rightSeparatorX] = '┬'
	}

	separator :=
		"├" +
			string(separatorChars) +
			"┤"

	content :=
		[]string{
			top,
			middle,
			separator,
		}

	panelLines :=
		strings.Split(
			fullPanels,
			"\n",
		)

	for _, line := range panelLines {

		padding :=
			m.width -
				lipgloss.Width(line) -
				2

		if padding < 0 {
			padding = 0
		}

		content =
			append(
				content,
				"│"+
					line+
					strings.Repeat(
						" ",
						padding,
					)+
					"│",
			)
	}

	bottomWidth :=
		m.width -
			2

	bottomChars :=
		[]rune(
			strings.Repeat(
				"─",
				bottomWidth,
			),
		)

	leftBottomX :=
		leftDividerX -
			1

	if leftBottomX >= 0 &&
		leftBottomX < len(bottomChars) {

		bottomChars[leftBottomX+1] = '┴'
	}

	rightBottomX :=
		rightDividerX -
			1

	if rightBottomX >= 0 &&
		rightBottomX < len(bottomChars) {

		bottomChars[rightBottomX] = '┴'
	}

	bottom :=
		"└" +
			string(bottomChars) +
			"┘"

	content =
		append(
			content,
			bottom,
		)

	return borderStyle.Render(
		strings.Join(
			content,
			"\n",
		),
	)
}
