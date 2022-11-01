package model

import (
	_ "embed"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	. "github.com/demostanis/metatorrent/internal/events"
	. "github.com/demostanis/metatorrent/internal/messages"
	. "github.com/demostanis/metatorrent/internal/torrent"
	. "github.com/demostanis/metatorrent/internal/views"
	"github.com/demostanis/metatorrent/providers"
	"net/url"
	"os"
	"os/exec"
	"sort"
)

type Model struct {
	torrents []Torrent
	errors   []string
	status   string

	query            textinput.Model
	torrentList      list.Model
	spinner          spinner.Model
	torrentListItems []list.Item

	terminalWidth  int
	terminalHeight int

	statusChannel   chan StatusMsg
	torrentsChannel chan TorrentsMsg
	errorsChannel   chan ErrorsMsg

	totalTorrentsCount     int
	processedTorrentsCount int
	loading                bool
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		WaitForStatus(m.statusChannel),
		WaitForTorrents(m.torrentsChannel),
		WaitForErrors(m.errorsChannel),
		m.spinner.Tick,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	getBodyHeight := func() int {
		return GetBodyHeight(TitleView(m.terminalWidth), ErrorView(m.errors, m.terminalWidth), m.query.View(), m.terminalHeight)
	}

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case StatusMsg:
		m.status = msg.Message
		if msg.IsLast {
			sort.Slice(m.torrents, func(i, j int) bool {
				return m.torrents[i].Seeders() > m.torrents[j].Seeders()
			})
			m.torrentListItems = torrentsToListItems(m.torrents)
			cmd = m.torrentList.SetItems(m.torrentListItems)
			m.status = fmt.Sprintf("Found %d torrents", m.processedTorrentsCount)
			m.torrentList.SetSize(m.terminalWidth, getBodyHeight())
			m.loading = false
		}
		return m, WaitForStatus(m.statusChannel)

	case TorrentsMsg:
		m.torrents = append(m.torrents, msg)
		m.processedTorrentsCount++

		var cmd tea.Cmd
		if len(m.torrents) != 0 {
			m.query.Reset()
			m.query.Blur()
		}

		return m, tea.Batch(cmd, WaitForTorrents(m.torrentsChannel))

	case ErrorsMsg:
		m.loading = false
		m.errors = append(m.errors, msg.Error())
		m.torrentList.SetSize(m.terminalWidth, getBodyHeight())
		return m, WaitForErrors(m.errorsChannel)

	case tea.KeyMsg:
		key := msg.String()
		// I wonder what quits the application...
		if key == "q" && m.query.Focused() {
			m.query, cmd = m.query.Update(msg)
			return m, cmd
		}
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if key == "ctrl+f" {
			m.query.Focus()
		}
		filterState := m.torrentList.FilterState().String()
		if filterState != "filtering" {
			if key == "esc" {
				return m, tea.Quit
			}
			if key == "enter" {
				m.torrents = make([]Torrent, 0)
				m.loading = true
				m.errors = make([]string, 0)
				m.processedTorrentsCount = 0
				go func() {
					query := url.QueryEscape(m.query.Value())
					providers.SearchWithEveryProvider(query, m.statusChannel, m.torrentsChannel, m.errorsChannel)
				}()
			} else if !m.query.Focused() && key == " " {
				myTorrent := m.torrentList.SelectedItem().(Torrent)
				magnet, err := myTorrent.Magnet()
				if err != nil {
					m.errors = append(m.errors, err.Error())
					return m, nil
				}

				torrentProgram := os.Getenv("TORRENT_PROGRAM")
				if torrentProgram == "" {
					torrentProgram = "transmission-gtk"
				}
				command := exec.Command(torrentProgram, magnet)
				err = command.Start()
				if err != nil {
					m.errors = append(m.errors, err.Error())
				}
			}
		}

	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		m.torrentListItems = torrentsToListItems(m.torrents)
		cmd = m.torrentList.SetItems(m.torrentListItems)
		m.torrentList.SetSize(m.terminalWidth, getBodyHeight())
	}

	cmds = append(cmds, cmd)

	m.query, cmd = m.query.Update(msg)
	cmds = append(cmds, cmd)

	m.torrentList, cmd = m.torrentList.Update(msg)
	cmds = append(cmds, cmd)

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return fmt.Sprintf("%s%s\n%s\n%s\n%s",
		TitleView(m.terminalWidth), StatusView(m.status, m.terminalWidth), ErrorView(m.errors, m.terminalWidth),
		BodyView(m.loading, m.errors, m.query, m.spinner, m.torrentList, m.terminalWidth, m.terminalHeight), m.query.View())
}

func makeTorrentList(torrents []list.Item) list.Model {
	torrentList := list.New(torrents, ItemDelegate{}, 0, 0)
	torrentList.SetShowHelp(false)
	torrentList.SetShowTitle(false)
	torrentList.SetShowHelp(false)
	torrentList.SetShowFilter(true)
	torrentList.SetFilteringEnabled(true)
	torrentList.FilterInput.CursorStyle = lipgloss.NewStyle()
	torrentList.FilterInput.PromptStyle = lipgloss.NewStyle()
	return torrentList
}

func makeTextInput() textinput.Model {
	input := textinput.New()
	input.Placeholder = "Search..."
	input.Focus()
	return input
}

func makeSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return s
}

func InitialModel() Model {
	torrentListItems := make([]list.Item, 0)

	torrentList := makeTorrentList(torrentListItems)
	input := makeTextInput()
	s := makeSpinner()

	return Model{
		torrentList: torrentList,
		query:       input,
		spinner:     s,

		statusChannel:   make(chan StatusMsg),
		torrentsChannel: make(chan TorrentsMsg),
		errorsChannel:   make(chan ErrorsMsg),

		torrentListItems: torrentListItems,
	}
}

func torrentsToListItems(torrents []Torrent) []list.Item {
	result := make([]list.Item, 0)
	for _, torrent := range torrents {
		result = append(result, list.Item(torrent))
	}
	return result
}
