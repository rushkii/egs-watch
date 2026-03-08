package main

import (
	"log"

	"github.com/rushkii/egs-watch/internal/config"
	"github.com/rushkii/egs-watch/internal/database"
	"github.com/rushkii/egs-watch/internal/epic"
	"github.com/rushkii/egs-watch/internal/repository"
	"github.com/rushkii/egs-watch/pkg"
)

func main() {
	http := pkg.NewClient()
	game := epic.NewEpicGames(http)

	db, err := database.New(config.PGHost, config.PGPort,
		config.PGUser, config.PGPwd, config.PGDb,
	)
	if err != nil {
		log.Println(err)
		return
	}

	defer db.Close()

	repo := repository.NewRepository(db)

	freeGames, err := game.GetFreeGamesFromEGS()
	if err != nil {
		log.Println(err)
		return
	}

	if err = repo.InsertFreeGames(freeGames.All); err != nil {
		log.Println(err)
		return
	}

	log.Println("Populate done")
}
