package messages

import (
	. "github.com/demostanis/metatorrent/internal/torrent"
)

type StatusMsg struct {
	Message string
	IsLast  bool
}
type TorrentsMsg Torrent
type ErrorsMsg error
