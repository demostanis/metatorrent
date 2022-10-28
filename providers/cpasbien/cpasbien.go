package providerCpasbien

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"sync"

	"github.com/antchfx/htmlquery"
	. "github.com/demostanis/metatorrent/internal/events"
	. "github.com/demostanis/metatorrent/internal/messages"
)

const Name = "cpasbien"
const MainUrl = "https://www.cpasbien.ch"

var providerCpasbienError = func(category, msg string) error {
	return errors.New(fmt.Sprintf("[%s/%s] ERROR: %s", Name, category, msg))
}

func searchPage(query string, beginning int, statusChannel chan StatusMsg, torrentsChannel chan TorrentsMsg) error {
	var wg sync.WaitGroup
	doc, err := htmlquery.LoadURL(fmt.Sprintf("%s/recherche/%s/%d", MainUrl, query, beginning))
	if err != nil {
		return providerCpasbienError("other", err.Error())
	}

	SendStatus(statusChannel, "[%s] Processing page %d...", Name, int(math.Floor(float64(beginning)/50)+1))

	titleElements, _ := htmlquery.QueryAll(doc, "//td/a[@class=\"titre\"]/text()")
	linkElements, _ := htmlquery.QueryAll(doc, "//td/a[@class=\"titre\"]/@href")
	sizeElements, _ := htmlquery.QueryAll(doc, "//td/div[@class=\"poid\"]")
	seedersCountElements, _ := htmlquery.QueryAll(doc, "//td/div[@class=\"down\"]")
	leechersCountElements, _ := htmlquery.QueryAll(doc, "//td/div[@class=\"up\"]//text()")

	if len(linkElements) != len(titleElements) ||
		len(seedersCountElements) != len(titleElements) ||
		len(leechersCountElements) != len(titleElements) ||
		len(sizeElements) != len(titleElements) {
		return providerCpasbienError("parsing", "Torrent entries are malformed.")
	}

	for i, title := range titleElements {
		title := htmlquery.InnerText(title)
		link := htmlquery.SelectAttr(linkElements[i], "href")

		seeders, err := strconv.Atoi(htmlquery.InnerText(seedersCountElements[i]))
		if err != nil {
			return providerCpasbienError("parsing", "Expected seeders to be a number.")
		}
		leechers, err := strconv.Atoi(htmlquery.InnerText(leechersCountElements[i]))
		if err != nil {
			return providerCpasbienError("parsing", "Expected leechers to be a number.")
		}
		size := htmlquery.InnerText(sizeElements[i])

		myTorrent := ProviderCpasbienTorrent{
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
	doc, err := htmlquery.LoadURL(fmt.Sprintf("%s/recherche/%s", MainUrl, query))
	if err != nil {
		errorsChannel <- err
		return
	}

	pages := htmlquery.Find(doc, "//ul[@class=\"pagination\"]//a")
	if len(pages) == 0 {
		errorsChannel <- providerCpasbienError("404", "No results found.")
		return
	}
	indexes := make([]int, 0)
	for _, page := range pages {
		matches := regexp.MustCompile(`\[(\d+)-\d+\]`).FindStringSubmatch(htmlquery.InnerText(page))
		if len(matches) > 0 {
			beginning, _ := strconv.Atoi(matches[1])
			indexes = append(indexes, beginning)
		}
	}
	pageCount := len(indexes)

	SendStatus(statusChannel, "[%s] Found %d pages", Name, pageCount)

	var lastError error
	var wg sync.WaitGroup

	for i := 0; i < pageCount; i++ {
		wg.Add(1)
		go func(i int) {
			err := searchPage(query, indexes[i], statusChannel, torrentsChannel)
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
