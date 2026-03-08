package repository

import (
	"fmt"
	"strings"
)

const BORDER = "-----------------------"

type sectionResult struct {
	Type    string
	Content string
}

func FormatFreeAllGames(result FreeGamesFilteredFromDB) (string, string) {
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

func formatFreeNowGames(games []FreeGamesFromDB) string {
	var sb strings.Builder

	sb.WriteString("```--------- NOW ---------\n")

	for index, res := range games {
		if res.OfferType != "BASE_GAME" {
			continue
		}

		endDate := res.EndDate.Local()

		sb.WriteString(fmt.Sprintf("Title     : %s\n", res.Title))
		sb.WriteString(fmt.Sprintf("Developer : %s\n", res.Developer))
		sb.WriteString(fmt.Sprintf("Publisher : %s\n", res.Publisher))
		sb.WriteString(fmt.Sprintf("Free Now  : Until %02d %s\n", endDate.Day(), endDate.Month()))

		if len(res.Slug) > 0 {
			sb.WriteString(fmt.Sprintf("https://store.epicgames.com/en-US/p/%s\n", res.Slug))
		}

		if index != len(games)-1 {
			sb.WriteString(BORDER + "\n")
		}
	}

	sb.WriteString("----- END OF NOW ------```\n")

	return sb.String()
}

func formatFreeUpcomingGames(games []FreeGamesFromDB) string {
	var sb strings.Builder

	sb.WriteString("```------ UPCOMING -------\n")

	for _, res := range games {
		if res.OfferType != "BASE_GAME" {
			continue
		}

		startDate := res.StartDate.Local()
		endDate := res.EndDate.Local()

		sb.WriteString(
			fmt.Sprintf("- %s (%02d %s - %02d %s)\n",
				res.Title,
				startDate.Day(),
				startDate.Month(),
				endDate.Day(),
				endDate.Month(),
			),
		)
	}

	sb.WriteString("--- END OF UPCOMING ---```\n")

	return sb.String()
}

func processSection(section string, games []FreeGamesFromDB, ch chan<- sectionResult) {
	var content string

	switch section {
	case "NOW":
		content = formatFreeNowGames(games)
	case "UPCOMING":
		content = formatFreeUpcomingGames(games)
	}

	ch <- sectionResult{Type: section, Content: content}
}

func GetImageWide(images []FGDBKeyImage) string {
	for _, img := range images {
		if img.Type == "OfferImageWide" {
			return img.URL
		}
	}
	return ""
}
