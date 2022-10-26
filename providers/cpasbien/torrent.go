package providerCpasbien

import (
	"github.com/antchfx/htmlquery"
)

type ProviderCpasbienTorrent struct {
	title    string
	link     string
	seeders  int
	leechers int
	size     string
}

func (t ProviderCpasbienTorrent) Title() string {
	return t.title
}

func (t ProviderCpasbienTorrent) Seeders() int {
	return t.seeders
}

func (t ProviderCpasbienTorrent) Leechers() int {
	return t.leechers
}

func (t ProviderCpasbienTorrent) Size() string {
	return t.size
}

func (t ProviderCpasbienTorrent) FilterValue() string {
	return t.title
}

func (t ProviderCpasbienTorrent) Magnet() (string, error) {
	doc, err := htmlquery.LoadURL(MainUrl + t.link)
	if err != nil {
		return "", err
	}
	magnetLink := htmlquery.FindOne(doc, "//div[@class=\"btn-download\"]/a/@href")
	if magnetLink == nil {
		return "", providerCpasbienError("parsing", "Download link is missing.")
	}
	href := htmlquery.SelectAttr(magnetLink, "href")
	return href, nil
}
