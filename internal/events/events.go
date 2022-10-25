package events

import (
	tea "github.com/charmbracelet/bubbletea"
	. "github.com/demostanis/metatorrent/internal/messages"
)

func WaitForStatus(statusChannel chan string) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg(<-statusChannel)
	}
}

func WaitForTorrents(torrentsChannel chan TorrentsMsg) tea.Cmd {
	return func() tea.Msg {
		return TorrentsMsg(<-torrentsChannel)
	}
}

func WaitForErrors(errorsChannel chan error) tea.Cmd {
	return func() tea.Msg {
		return ErrorsMsg(<-errorsChannel)
	}
}
