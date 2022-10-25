package providerPirateBay

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/demostanis/metatorrent/internal/messages"
	"io"
	"net/http"
	"strconv"
	"sync"
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

func Search(query string, statusChannel chan StatusMsg, torrentsChannel chan TorrentsMsg, errorsChannel chan ErrorsMsg) {
	var wg sync.WaitGroup
	status(statusChannel, false, "[%s] Processing...", Name)

	resp, err := http.Get(fmt.Sprintf("%s/q.php?q=%s", MainUrl, query))
	if err != nil {
		errorsChannel <- err
		return
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

	for _, entry := range data {
		if entry.Name == "No results returned" {
			continue
		}

		seeders, err := strconv.Atoi(entry.Seeders)
		if err != nil {
			errorsChannel <- providerPirateBayError("parsing", "Expected seeders to be a number.")
			return
		}
		leechers, err := strconv.Atoi(entry.Leechers)
		if err != nil {
			errorsChannel <- providerPirateBayError("parsing", "Expected leechers to be a number.")
			return
		}
		size, err := strconv.ParseUint(entry.Size, 10, 64)
		if err != nil {
			errorsChannel <- providerPirateBayError("parsing", "Expected size to be a number.")
			return
		}

		myTorrent := ProviderPirateBayTorrent{
			title:    entry.Name,
			hash:     entry.Hash,
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
	status(statusChannel, true, "[%s] Processed...", Name)
	return
}

func status(statusChannel chan StatusMsg, isLast bool, message string, rest ...any) {
	go func() {
		statusChannel <- StatusMsg{fmt.Sprintf(message, rest...), isLast}
	}()
}
