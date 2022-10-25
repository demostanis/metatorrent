package providerPirateBay

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	. "github.com/demostanis/metatorrent/internal/messages"
)

const Name = "The Pirate Bay"
const MainUrl = "https://apibay.org"

var providerPirateBayError = func(category, msg string) error {
	return errors.New(fmt.Sprintf("[%s/%s] ERROR: %s", Name, category, msg))
}

type entry struct {
	Leechers string
	Seeders  string
	Name     string
	Size     string
	Hash     string `json:"info_hash"`
}

func Search(query string, statusChannel chan string, torrentsChannel chan TorrentsMsg, errorsChannel chan error) {
	status(statusChannel, "[%s] Processing...", Name)

	resp, err := http.Get(fmt.Sprintf("%s/q.php?q=%s", MainUrl, query))
	if err != nil {
		errorsChannel <- err
	}
	defer resp.Body.Close()

	data := make([]entry, 0)
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		errorsChannel <- err
	}
	err = json.Unmarshal(b, &data)
	if err != nil {
		errorsChannel <- err
	}

	for i, entry := range data {
		if entry.Name == "No results returned" {
			continue
		}

		seeders, err := strconv.Atoi(entry.Seeders)
		if err != nil {
			errorsChannel <- providerPirateBayError("parsing", "Expected seeders to be a number.")
		}
		leechers, err := strconv.Atoi(entry.Leechers)
		if err != nil {
			errorsChannel <- providerPirateBayError("parsing", "Expected leechers to be a number.")
		}
		size, err := strconv.ParseUint(entry.Size, 10, 64)
		if err != nil {
			errorsChannel <- providerPirateBayError("parsing", "Expected size to be a number.")
		}

		last := false
		if i == len(data)-1 {
			last = true
		}
		myTorrent := TorrentsMsg{
			ProviderPirateBayTorrent{
				title:    entry.Name,
				hash:     entry.Hash,
				seeders:  seeders,
				leechers: leechers,
				size:     size,
			},
			last,
		}
		go func() {
			torrentsChannel <- myTorrent
		}()
	}
}

func status(statusChannel chan string, message string, rest ...any) {
	go func() {
		statusChannel <- fmt.Sprintf(message, rest...)
	}()
}
