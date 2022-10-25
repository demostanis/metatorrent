// A simple package to aggregate all Torrent providers.

package providers

import (
	. "github.com/demostanis/metatorrent/internal/messages"
	"github.com/demostanis/metatorrent/providers/1337x"
	"github.com/demostanis/metatorrent/providers/thepiratebay"
)

func SearchWithEveryProvider(query string, statusChannel chan string,
	torrentsChannel chan TorrentsMsg, errorsChannel chan error) {
	provider1337x.Search(query, statusChannel, torrentsChannel, errorsChannel)
	providerPirateBay.Search(query, statusChannel, torrentsChannel, errorsChannel)
}

func SupportedProviders() []string {
	return []string{
		provider1337x.Name,
		providerPirateBay.Name,
	}
}
