package epic

import "github.com/rushkii/egs-watch/pkg"

type FreeGamesFilter int

const (
	FilterAll FreeGamesFilter = 1 << iota
	FilterFreeNow
	FilterUpcoming
)

type EpicGames struct {
	client          *pkg.HttpClient
	FILTER_ALL      FreeGamesFilter
	FILTER_FREE_NOW FreeGamesFilter
	FILTER_UPCOMING FreeGamesFilter
}

func NewEpicGames(client *pkg.HttpClient) *EpicGames {
	return &EpicGames{
		client:          client,
		FILTER_ALL:      FilterAll,
		FILTER_FREE_NOW: FilterFreeNow,
		FILTER_UPCOMING: FilterUpcoming,
	}
}
