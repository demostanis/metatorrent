package providerPirateBay

import (
	"fmt"
	"net/url"

	"github.com/dustin/go-humanize"
)

type ProviderPirateBayTorrent struct {
	title    string
	seeders  int
	leechers int
	size     uint64
	hash     string
}

func (t ProviderPirateBayTorrent) Title() string {
	return t.title
}

func (t ProviderPirateBayTorrent) Seeders() int {
	return t.seeders
}

func (t ProviderPirateBayTorrent) Leechers() int {
	return t.leechers
}

func (t ProviderPirateBayTorrent) Size() string {
	return humanize.Bytes(t.size)
}

func (t ProviderPirateBayTorrent) FilterValue() string {
	return t.title
}

func (t ProviderPirateBayTorrent) Magnet() (string, error) {
	trackers := []string{
		"udp://185.193.125.139:6969/announce",
		"udp://tracker.opentrackr.org:1337",
		"udp://tracker.openbittorrent.com:6969/announce",
		"udp://movies.zsw.ca:6969/announce",
		"udp://open.stealth.si:80/announce",
		"udp://tracker.0x.tf:6969/announce",
		"udp://opentracker.i2p.rocks:6969/announce",
		"udp://tracker.tiny-vps.com:6969/announce",
		"udp://tracker.torrent.eu.org:451/announce",
		"udp://tracker.internetwarriors.net:1337/announce",
		"udp://tracker.dler.org:6969/announce"}

	trackersStr := ""
	for _, tracker := range trackers {
		trackersStr += "&tr=" + url.QueryEscape(tracker)
	}

	return fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s&%s", t.hash, url.QueryEscape(t.title), trackersStr), nil
}
