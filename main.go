package main

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/demostanis/metatorrent/internal"
	"github.com/demostanis/metatorrent/providers/1337x"
	"os"
)

func searchWithEveryProvider(query string) []torrents.Torrent {
	finalTorrents := new([]torrents.Torrent)
	err, torrents := provider1337x.Search(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Searching torrents with 1337x failed: %s\n", err.Error())
	} else {
		for _, torrent := range torrents {
			*finalTorrents = append(*finalTorrents, torrent)
		}
	}
	if len(*finalTorrents) == 0 {
		fmt.Fprintln(os.Stderr, "Could not find any torrents.")
	}
	return *finalTorrents
}

func main() {
	rl, err := readline.New("> ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	line, err := rl.Readline()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	torrents := searchWithEveryProvider(line)
	for _, torrent := range torrents {
		fmt.Printf("%s | %d seeders, %d leechers | %s\n", torrent.Title, torrent.Seeders, torrent.Leechers, torrent.Size)
	}
}
