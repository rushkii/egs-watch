package epic

const API = "https://store-site-backend-static-ipv4.ak.epicgames.com/freeGamesPromotions?locale=en-US&country=ID&allowCountries=ID"

func (c *EpicGames) GetFreeGamesFromEGS() (FreeGamesFiltered, error) {
	var response FreeGamesResponse

	if err := c.client.GetJSON(API, &response); err != nil {
		return FreeGamesFiltered{}, err
	}

	var filtered FreeGamesFiltered

	filtered.All = response.Data.Catalog.SearchStore.Elements

	for _, el := range response.Data.Catalog.SearchStore.Elements {
		if isFreeNow(el) {
			filtered.Now = append(filtered.Now, el)
		} else if isFreeUpcoming(el) {
			filtered.Upcoming = append(filtered.Upcoming, el)
		}
	}

	return filtered, nil
}
