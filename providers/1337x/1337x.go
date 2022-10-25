package provider1337x

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/antchfx/htmlquery"
	. "github.com/demostanis/metatorrent/internal/messages"
)

const Name = "1337x"
const MainUrl = "https://www.1337x.to"

var provider1337xError = func(category, msg string) error {
	return errors.New(fmt.Sprintf("[%s/%s] ERROR: %s", Name, category, msg))
}

func searchPage(query string, page int, lastPage int, statusChannel chan StatusMsg, torrentsChannel chan TorrentsMsg) error {
	var wg sync.WaitGroup
	doc, err := htmlquery.LoadURL(fmt.Sprintf("%s/search/%s/%d/", MainUrl, query, page))
	if err != nil {
		return err
	}

	status(statusChannel, false, "[%s] Processing page %d...", Name, page)

	titleElements, err := htmlquery.QueryAll(doc, "//td[@class=\"coll-1 name\"]/a[2]")
	linkElements, err := htmlquery.QueryAll(doc, "//td[@class=\"coll-1 name\"]//a[2]/@href")
	seedersCountElements, err := htmlquery.QueryAll(doc, "//td[@class=\"coll-2 seeds\"]")
	leechersCountElements, err := htmlquery.QueryAll(doc, "//td[@class=\"coll-3 leeches\"]")
	sizeElements, err := htmlquery.QueryAll(doc, "//td[contains(@class, \"coll-4 size\")]/text()")

	// Since the elements are in a table, they should always be in identical counts.
	// In case they're not, we assume there was a parsing error, and bail out.
	if len(linkElements) != len(titleElements) ||
		len(seedersCountElements) != len(titleElements) ||
		len(leechersCountElements) != len(titleElements) ||
		len(sizeElements) != len(titleElements) {
		return provider1337xError("parsing", "Torrent entries are malformed.")
	}

	for i, title := range titleElements {
		title := htmlquery.InnerText(title)
		link := htmlquery.SelectAttr(linkElements[i], "href")

		seeders, err := strconv.Atoi(htmlquery.InnerText(seedersCountElements[i]))
		if err != nil {
			return provider1337xError("parsing", "Expected seeders to be a number.")
		}
		leechers, err := strconv.Atoi(htmlquery.InnerText(leechersCountElements[i]))
		if err != nil {
			return provider1337xError("parsing", "Expected leechers to be a number.")
		}
		size := htmlquery.InnerText(sizeElements[i])
		if err != nil {
			return provider1337xError("parsing", "Torrent size is missing.")
		}

		myTorrent := Provider1337xTorrent{
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

	wg.Wait()
	isLast := false
	if page == lastPage {
		isLast = true
	}
	status(statusChannel, isLast, "[%s] Processed page %d...", Name, page)
	return nil
}

// Finds the number of pages of results for `query`, and scrapes all of them using `searchPage`.
func Search(query string, statusChannel chan StatusMsg, torrentsChannel chan TorrentsMsg, errorsChannel chan ErrorsMsg) {
	doc, err := htmlquery.LoadURL(fmt.Sprintf("%s/search/%s/1/", MainUrl, query))
	if err != nil {
		errorsChannel <- err
		return
	}

	lastPageElem := htmlquery.FindOne(doc, "//div[@class=\"pagination\"]//li[last()]//@href")
	if lastPageElem == nil {
		errorsChannel <- provider1337xError("parsing", "Max page number is missing.")
		return
	}
	href := htmlquery.SelectAttr(lastPageElem, "href")
	match := regexp.MustCompile("/(\\d+)/").FindString(href)
	lastPage, err := strconv.Atoi(match[1 : len(match)-1])
	if err != nil {
		errorsChannel <- err
		return
	}
	status(statusChannel, false, "[%s] Found %d pages", Name, lastPage)

	var lastError error
	scrapedPages := 0

	for i := 1; i <= lastPage; i++ {
		go func(i int) {
			err := searchPage(query, i, lastPage, statusChannel, torrentsChannel)
			if err != nil {
				lastError = err
				return
			}
			scrapedPages++
		}(i)
	}
	for {
		if lastError != nil || scrapedPages == lastPage {
			break
		}
	}

	if lastError != nil {
		errorsChannel <- lastError
		return
	}
}

// The status channel is also used to inform all torrents have been sent, hence `isLast`.
func status(statusChannel chan StatusMsg, isLast bool, message string, rest ...any) {
	go func() {
		statusChannel <- StatusMsg{fmt.Sprintf(message, rest...), isLast}
	}()
}
