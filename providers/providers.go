// A simple package to aggregate all Torrent providers.

package providers

import (
	. "github.com/demostanis/metatorrent/internal/messages"
	"github.com/demostanis/metatorrent/providers/1337x"
	"github.com/demostanis/metatorrent/providers/cpasbien"
	"github.com/demostanis/metatorrent/providers/thepiratebay"
	"github.com/demostanis/metatorrent/providers/yggtorrent"
)

func SearchWithEveryProvider(query string, statusChannel chan StatusMsg,
	torrentsChannel chan TorrentsMsg, errorsChannel chan ErrorsMsg) {
	go provider1337x.Search(query, statusChannel, torrentsChannel, errorsChannel)
	go providerPirateBay.Search(query, statusChannel, torrentsChannel, errorsChannel)
	go providerCpasbien.Search(query, statusChannel, torrentsChannel, errorsChannel)
	go providerYggTorrent.Search(query, statusChannel, torrentsChannel, errorsChannel)
}

func SupportedProviders() []string {
	return []string{
		provider1337x.Name,
		providerPirateBay.Name,
		providerCpasbien.Name,
		providerYggTorrent.Name,
	}
}
