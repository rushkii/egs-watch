package epic

import (
	"fmt"
	"strings"

	"github.com/rushkii/egs-watch/pkg"
)

const BORDER = "-----------------------"

type sectionResult struct {
	Type    string
	Content string
}

func FormatFreeAllGames(result FreeGamesFiltered) (string, string) {
	resultsChan := make(chan sectionResult, 2)

	go processSection("NOW", result.Now, resultsChan)
	go processSection("UPCOMING", result.Upcoming, resultsChan)

	var now, upcoming string

	for i := 0; i < 2; i++ {
		msg := <-resultsChan
		switch msg.Type {
		case "NOW":
			now = msg.Content
		case "UPCOMING":
			upcoming = msg.Content
		}
	}

	return now, upcoming
}

func formatFreeNowGames(games []FGElement) string {
	var sb strings.Builder

	sb.WriteString("```--------- NOW ---------\n")

	for index, res := range games {
		if res.OfferType != "BASE_GAME" || res.Promotions == nil {
			continue
		}

		attrs := make([]pkg.KeyValue, len(res.CustomAttributes))

		for i, a := range res.CustomAttributes {
			attrs[i] = pkg.KeyValue{Key: a.Key, Value: a.Value}
		}

		developer := pkg.GetKVFromArray("developerName", attrs)
		publisher := pkg.GetKVFromArray("publisherName", attrs)

		if developer == "" {
			developer = res.Seller.Name
		}

		if publisher == "" {
			publisher = res.Seller.Name
		}

		endDate := res.Promotions.PromotionalOffers[0].PromotionalOffers[0].EndDate.Local()

		sb.WriteString(fmt.Sprintf("Title     : %s\n", res.Title))
		sb.WriteString(fmt.Sprintf("Developer : %s\n", developer))
		sb.WriteString(fmt.Sprintf("Publisher : %s\n", publisher))
		sb.WriteString(fmt.Sprintf("Free Now  : Until %02d %s\n", endDate.Day(), endDate.Month()))

		if len(res.CatalogNs.Mappings) > 0 {
			sb.WriteString(fmt.Sprintf("https://store.epicgames.com/en-US/p/%s\n", res.CatalogNs.Mappings[0].PageSlug))
		}

		if index != len(games)-1 {
			sb.WriteString(BORDER + "\n")
		}
	}

	sb.WriteString("----- END OF NOW ------```\n")

	return sb.String()
}

func formatFreeUpcomingGames(games []FGElement) string {
	var sb strings.Builder

	sb.WriteString("```------ UPCOMING -------\n")

	for _, res := range games {
		if res.OfferType != "BASE_GAME" || res.Promotions == nil {
			continue
		}

		startDate := res.Promotions.UpcomingPromotionalOffers[0].PromotionalOffers[0].StartDate.Local()
		endDate := res.Promotions.UpcomingPromotionalOffers[0].PromotionalOffers[0].EndDate.Local()

		sb.WriteString(fmt.Sprintf("- %s (%02d %s - %02d %s)\n", res.Title, startDate.Day(), startDate.Month(), endDate.Day(), endDate.Month()))
	}

	sb.WriteString("--- END OF UPCOMING ---```\n")

	return sb.String()
}

func processSection(section string, games []FGElement, ch chan<- sectionResult) {
	var content string

	switch section {
	case "NOW":
		content = formatFreeNowGames(games)
	case "UPCOMING":
		content = formatFreeUpcomingGames(games)
	}

	ch <- sectionResult{Type: section, Content: content}
}

func isFreeNow(el FGElement) bool {
	return el.Promotions != nil &&
		len(el.Promotions.PromotionalOffers) > 0 &&
		len(el.Promotions.PromotionalOffers[0].PromotionalOffers) > 0
}

func isFreeUpcoming(el FGElement) bool {
	return el.Promotions != nil &&
		len(el.Promotions.UpcomingPromotionalOffers) > 0 &&
		len(el.Promotions.UpcomingPromotionalOffers[0].PromotionalOffers) > 0
}

func GetImageWide(images []FGKeyImage) string {
	for _, img := range images {
		if img.Type == "OfferImageWide" {
			return img.URL
		}
	}
	return ""
}
