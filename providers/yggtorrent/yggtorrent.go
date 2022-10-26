package providerYggTorrent

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
	. "github.com/demostanis/metatorrent/internal/events"
	. "github.com/demostanis/metatorrent/internal/messages"
)

const Name = "YggTorrent"
const MainUrl = "https://www2.yggtorrent.co"

var providerYggTorrentError = func(category, msg string) error {
	return errors.New(fmt.Sprintf("[%s/%s] ERROR: %s", Name, category, msg))
}

func searchPage(query string, page int, statusChannel chan StatusMsg, torrentsChannel chan TorrentsMsg) error {
	var wg sync.WaitGroup
	doc, err := htmlquery.LoadURL(fmt.Sprintf("%s/search_torrent/%s/page-%d", MainUrl, query, page))
	if err != nil {
		return err
	}

	SendStatus(statusChannel, "[%s] Processing page %d...", Name, page)

	titleElements, _ := htmlquery.QueryAll(doc, "//td//a/@title")
	linkElements, _ := htmlquery.QueryAll(doc, "//td/a/@href")
	sizeElements, _ := htmlquery.QueryAll(doc, "//tr/td[2]")
	seedersCountElements, _ := htmlquery.QueryAll(doc, "//tr/td[3]/span")
	leechersCountElements, _ := htmlquery.QueryAll(doc, "//tr/td[4]")

	if len(linkElements) != len(titleElements) ||
		len(seedersCountElements) != len(titleElements) ||
		len(leechersCountElements) != len(titleElements) ||
		len(sizeElements) != len(titleElements) {
		return providerYggTorrentError("parsing", "Torrent entries are malformed.")
	}

	for i, title := range titleElements {
		title := htmlquery.SelectAttr(title, "title")
		link := htmlquery.SelectAttr(linkElements[i], "href")

		seeders, err := strconv.Atoi(strings.TrimSpace(htmlquery.InnerText(seedersCountElements[i])))
		if err != nil {
			return providerYggTorrentError("parsing", "Expected seeders to be a number.")
		}
		leechers, err := strconv.Atoi(strings.TrimSpace(htmlquery.InnerText(leechersCountElements[i])))
		if err != nil {
			return providerYggTorrentError("parsing", "Expected leechers to be a number.")
		}
		size := htmlquery.InnerText(sizeElements[i])

		myTorrent := ProviderYggTorrentTorrent{
			title:    title,
			link:     link,
			seeders:  seeders,
			leechers: leechers,
			size:     size,
		}
		wg.Add(1)
		go func() {
			torrentsChannel <- myTorrent
			wg.Done()
		}()
	}

	SendStatus(statusChannel, "[%s] Processed %d torrents...", Name, len(titleElements))
	wg.Wait()
	return nil
}

func Search(query string, statusChannel chan StatusMsg, torrentsChannel chan TorrentsMsg, errorsChannel chan ErrorsMsg) {
	doc, err := htmlquery.LoadURL(fmt.Sprintf("%s/search_torrent/%s.html", MainUrl, query))
	if err != nil {
		errorsChannel <- err
		return
	}

	pageCountElems := htmlquery.Find(doc, "//ul[@class=\"pagination\"]//a")
	if len(pageCountElems) == 0 {
		errorsChannel <- providerYggTorrentError("404", "No results found.")
		return
	}
	pageCountElem := pageCountElems[len(pageCountElems)-2]
	if pageCountElem == nil {
		errorsChannel <- providerYggTorrentError("parsing", "Number of pages is missing.")
		return
	}
	pageCount, _ := strconv.Atoi(htmlquery.InnerText(pageCountElem))

	SendStatus(statusChannel, "[%s] Found %d pages", Name, pageCount)

	var lastError error
	var wg sync.WaitGroup

	for i := 1; i <= pageCount; i++ {
		wg.Add(1)
		go func(i int) {
			err := searchPage(query, i, statusChannel, torrentsChannel)
			wg.Done()
			if err != nil {
				lastError = err
				return
			}
		}(i)
	}

	wg.Wait()
	SendFinalStatus(statusChannel, "[%s] Done", Name)
	if lastError != nil {
		errorsChannel <- lastError
		return
	}
}
