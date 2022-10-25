package torrent

type Torrent interface {
	Title() string
	Seeders() int
	Leechers() int
	Size() string
	Magnet() (string, error)
	// This is used by bubbles' textinput
	FilterValue() string
}
