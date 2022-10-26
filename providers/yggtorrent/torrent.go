package providerYggTorrent

import (
	"github.com/antchfx/htmlquery"
)

type ProviderYggTorrentTorrent struct {
	title    string
	link     string
	seeders  int
	leechers int
	size     string
}

func (t ProviderYggTorrentTorrent) Title() string {
	return t.title
}

func (t ProviderYggTorrentTorrent) Seeders() int {
	return t.seeders
}

func (t ProviderYggTorrentTorrent) Leechers() int {
	return t.leechers
}

func (t ProviderYggTorrentTorrent) Size() string {
	return t.size
}

func (t ProviderYggTorrentTorrent) FilterValue() string {
	return t.title
}

func (t ProviderYggTorrentTorrent) Magnet() (string, error) {
	doc, err := htmlquery.LoadURL(MainUrl + t.link)
	if err != nil {
		return "", err
	}
	magnetLink := htmlquery.FindOne(doc, "//div[@class=\"btn-download\"]/a/@href")
	if magnetLink == nil {
		return "", providerYggTorrentError("parsing", "Download link is missing.")
	}
	href := htmlquery.SelectAttr(magnetLink, "href")
	return href, nil
}
