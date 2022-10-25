package messages

import (
	. "github.com/demostanis/metatorrent/internal/torrent"
)

type StatusMsg string
type TorrentsMsg struct {
	Torrent Torrent
	Last    bool
}
type ErrorsMsg error
