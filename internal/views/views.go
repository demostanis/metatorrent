package views

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	. "github.com/demostanis/metatorrent/internal/logo"
	. "github.com/demostanis/metatorrent/internal/messages"
	. "github.com/demostanis/metatorrent/internal/torrent"
	"github.com/mattn/go-runewidth"
	"io"
	"strings"
)

var (
	centerStyle = func(w int) lipgloss.Style {
		return lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(w)
	}

	errorStyle = func(w int) lipgloss.Style {
		return centerStyle(w).
			Foreground(lipgloss.Color("#ff4545")).
			Background(lipgloss.Color("#363636")).
			MarginTop(1).
			Padding(3).
			Bold(true).
			Width(w)
	}

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#44D3F7")).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true)
)

func StatusView(status StatusMsg, w int) string {
	return centerStyle(w).
		Render(status.Message)
}

func TitleView(w int) string {
	s := "\n"
	for _, line := range strings.Split(Logo, "\n") {
		s += centerStyle(w).
			Inline(true).
			Render(line) + "\n"
	}
	return s
}

func ErrorView(err error, w int) string {
	if err != nil {
		return errorStyle(w).
			Render(err.Error())
	}
	return ""
}

func WelcomeScreenView(err error, query textinput.Model, w int, h int) string {
	return centerStyle(w).
		Height(GetBodyHeight(
			TitleView(w), ErrorView(err, w),
			query.View(), h,
		)).
		// TODO: Add more colors
		Render("\n" + `Write your search and press Enter to start.
You can filter the results using /, scroll with j and k,
download a torrent by pressing space and make a new search by using C-f.`)
}

// Too bad we can't pass a Model here (due to cyclic imports)
func BodyView(loading bool, err error, query textinput.Model,
	spinner spinner.Model, torrentList list.Model, w int, h int) string {
	body := centerStyle(w).
		Height(GetBodyHeight(
			TitleView(w), ErrorView(err, w),
			query.View(), h,
		)).
		Render(spinner.View())

	if !loading {
		if len(torrentList.Items()) != 0 {
			body = torrentList.View()
		} else {
			body = WelcomeScreenView(err, query, w, h)
		}
	}

	return body
}

// We make our own delegate function to render each element
// of the list ourselves. Here we make the currently selected
// one bold.
type ItemDelegate struct{}

func (d ItemDelegate) Height() int                               { return 1 }
func (d ItemDelegate) Spacing() int                              { return 0 }
func (d ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	t, ok := item.(Torrent)

	if !ok {
		return
	}

	style := lipgloss.NewStyle()
	if index == m.Index() {
		style = selectedStyle
	}

	rightMost := fmt.Sprintf("%d↑, %d↓ | %s",
		t.Seeders(), t.Leechers(), t.Size())

	properTitle := t.Title()
	if m.Width() < len(t.Title())+len(rightMost) {
		properTitle = t.Title()[:m.Width()-len(rightMost)-3] + "…"
	}
	spacing := strings.Repeat(" ", m.Width()-runewidth.StringWidth(properTitle)-len(rightMost)+4)

	fmt.Fprintf(w, style.Render(properTitle+spacing+rightMost))
}

func GetBodyHeight(titleView string, errorView string, inputView string, terminalHeight int) int {
	headerSize := lipgloss.Height(titleView) + lipgloss.Height(errorView)
	footerSize := lipgloss.Height(inputView)
	return terminalHeight - headerSize - footerSize
}
