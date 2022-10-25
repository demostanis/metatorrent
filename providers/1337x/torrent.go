package provider1337x

import (
	"github.com/antchfx/htmlquery"
)

type Provider1337xTorrent struct {
	title    string
	link     string
	seeders  int
	leechers int
	size     string
}

func (t Provider1337xTorrent) Title() string {
	return t.title
}

func (t Provider1337xTorrent) Seeders() int {
	return t.seeders
}

func (t Provider1337xTorrent) Leechers() int {
	return t.leechers
}

func (t Provider1337xTorrent) Size() string {
	return t.size
}

func (t Provider1337xTorrent) FilterValue() string {
	return t.title
}

func (t Provider1337xTorrent) Magnet() (string, error) {
	doc, err := htmlquery.LoadURL(MainUrl + t.link)
	if err != nil {
		return "", err
	}
	magnetLink := htmlquery.FindOne(doc, "//a[starts-with(@href, \"magnet:\")]/@href")
	if magnetLink == nil {
		return "", provider1337xError("parsing", "Magnet link is missing.")
	}
	href := htmlquery.SelectAttr(magnetLink, "href")
	return href, nil
}
