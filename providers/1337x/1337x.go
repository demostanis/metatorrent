package provider1337x

import (
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/demostanis/metatorrent/internal"
	"regexp"
	"sort"
	"strconv"
)

const Name = "1337x"
const MainUrl = "https://www.1337x.to"

var elementsMissingError = errors.New("Did not find expected elements on the page")
var elementsInvalidError = errors.New("Found elements on the page, but with wrong values")

func searchPage(query string, page int) (error, []torrents.Torrent) {
	doc, err := htmlquery.LoadURL(fmt.Sprintf("%s/search/%s/%d/", MainUrl, query, page))
	if err != nil {
		return err, nil
	}

	fmt.Printf("Processing page %d...\n", page)

	titles, err := htmlquery.QueryAll(doc, "//td[@class=\"coll-1 name\"]/a[2]")
	seeders, err := htmlquery.QueryAll(doc, "//td[@class=\"coll-2 seeds\"]")
	leechers, err := htmlquery.QueryAll(doc, "//td[@class=\"coll-3 leeches\"]")
	sizes, err := htmlquery.QueryAll(doc, "//td[contains(@class, \"coll-4 size\")]/text()")

	var myTorrents []torrents.Torrent

	if len(seeders) != len(titles) || len(leechers) != len(titles) || len(sizes) != len(titles) {
		return elementsMissingError, nil
	}

	for i, title := range titles {
		titleValue := htmlquery.InnerText(title)

		seedersValue, err := strconv.Atoi(htmlquery.InnerText(seeders[i]))
		if err != nil {
			return elementsInvalidError, nil
		}
		leechersValue, err := strconv.Atoi(htmlquery.InnerText(leechers[i]))
		if err != nil {
			return elementsInvalidError, nil
		}
		sizeValue := htmlquery.InnerText(sizes[i])
		if err != nil {
			return elementsInvalidError, nil
		}

		myTorrents = append(myTorrents, torrents.Torrent{
			Title:    titleValue,
			Seeders:  seedersValue,
			Leechers: leechersValue,
			Size:     sizeValue,
		})
	}

	return nil, myTorrents
}

func Search(query string) (error, []torrents.Torrent) {
	doc, err := htmlquery.LoadURL(fmt.Sprintf("%s/search/%s/1/", MainUrl, query))
	if err != nil {
		return err, nil
	}
	lastPageElem := htmlquery.FindOne(doc, "//div[@class=\"pagination\"]//li[last()]//@href")
	if lastPageElem == nil {
		return elementsMissingError, nil
	}

	href := htmlquery.SelectAttr(lastPageElem, "href")
	match := regexp.MustCompile("/(\\d+)/").FindString(href)
	lastPage, err := strconv.Atoi(match[1 : len(match)-1])
	if err != nil {
		return err, nil
	}
	fmt.Printf("Found %d pages\n", lastPage)

	var myTorrents []torrents.Torrent
	var lastError error
	done := 0
	for i := 1; i <= lastPage; i++ {
		go func(i int) {
			err, torrents := searchPage(query, i)
			if err != nil {
				lastError = err
				return
			}
			for _, torrent := range torrents {
				myTorrents = append(myTorrents, torrent)
			}
			done++
		}(i)
	}
	for {
		if lastError != nil || done == lastPage {
			break
		}
	}
	if lastError != nil {
		return lastError, nil
	}
	sort.Slice(myTorrents, func(i, j int) bool {
		return myTorrents[i].Seeders > myTorrents[j].Seeders
	})
	return nil, myTorrents
}
