package main

import (
	"encoding/json"
	"log"

	"github.com/rushkii/egs-watch/internal/config"
	"github.com/rushkii/egs-watch/internal/database"
	"github.com/rushkii/egs-watch/internal/epic"
	"github.com/rushkii/egs-watch/internal/repository"
	"github.com/rushkii/egs-watch/pkg"
)

func PrintFreeGamesAllFromDB() {
	db, err := database.New(config.PGHost, config.PGPort,
		config.PGUser, config.PGPwd, config.PGDb,
	)
	if err != nil {
		log.Println(err)
		return
	}

	defer db.Close()

	repo := repository.NewRepository(db)

	result, err := repo.GetFreeGamesFromDB()
	if err != nil {
		log.Fatal(err)
		return
	}

	um, err := json.MarshalIndent(result.All, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(um))
}

func PrintFreeGamesAllFormatted() {
	game := epic.NewEpicGames(pkg.NewClient())
	result, _ := game.GetFreeGamesFromEGS()
	now, upcoming := epic.FormatFreeAllGames(result)

	log.Println(now)
	log.Println(upcoming)
}

func main() {
	PrintFreeGamesAllFromDB()
}
