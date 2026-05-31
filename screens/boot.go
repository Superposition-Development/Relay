package screens

import (
	"fmt"
	"math"
	"strings"
	"time"

	"Relay/app"
	_ "embed"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const pulseSpeed = 0.01
const pulseAmp = 0.1

type bootTickMsg struct{}

//go:embed boot.txt
var bootArt string

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
		return bootTickMsg{}
	})
}

type Phase int

const (
	Reveal Phase = iota
	Unwrite
)

type BootScreen struct {
	width  int
	height int

	frame int

	lines   []string
	visible int

	fade float64

	phase Phase

	holdTicks int
}

func NewBootScreen() BootScreen {
	return BootScreen{
		lines: strings.Split(bootArt, "\n"),
		fade:  0,
		phase: Reveal,
	}
}

func (b BootScreen) Init() tea.Cmd {
	return tick()
}

func (b BootScreen) Update(msg tea.Msg) (app.Screen, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		return b, func() tea.Msg {
			return app.ChangeScreenMsg{
				Screen: CreateModel(b.height, b.width),
				Width:  b.width,
				Height: b.height,
			}
		}

	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height

	case bootTickMsg:

		b.frame++

		switch b.phase {

		case Reveal:

			if b.visible < len(b.lines) {
				b.visible++
			} else {
				b.holdTicks++
				if b.fade < 1 {
					b.fade += 0.02
					if b.fade > 1 {
						b.fade = 1
					}
				}

				if b.holdTicks > 20 {
					b.phase = Unwrite
				}
			}

			if b.visible < len(b.lines) {
				if b.fade < 1 {
					b.fade += 0.04
					if b.fade > 1 {
						b.fade = 1
					}
				}
			}

		case Unwrite:

			if b.visible > 0 {
				b.visible--
			} else {
				return b, func() tea.Msg {
					return app.ChangeScreenMsg{
						Screen: CreateModel(b.height, b.width),
						Width:  b.width,
						Height: b.height,
					}
				}
			}

			if b.fade > 0 {
				b.fade -= 0.02
				if b.fade < 0 {
					b.fade = 0
				}
			}
		}

		return b, tick()
	}

	return b, nil
}

func (b BootScreen) View() string {

	pulse := 0.85 + pulseAmp*math.Sin(float64(b.frame)*pulseSpeed)
	intensity := b.fade * pulse

	style := lipgloss.NewStyle().Foreground(
		lipgloss.Color(fmt.Sprintf("#FF%02X00", int(140*intensity))),
	)

	var out []string

	for i := 0; i < b.visible && i < len(b.lines); i++ {
		out = append(out, style.Render(b.lines[i]))
	}

	content := strings.Join(out, "\n")

	return lipgloss.Place(
		b.width,
		b.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
