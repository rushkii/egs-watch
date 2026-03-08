package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rushkii/egs-watch/internal/config"
	"github.com/rushkii/egs-watch/internal/database"
	"github.com/rushkii/egs-watch/internal/epic"
	"github.com/rushkii/egs-watch/internal/repository"
	"github.com/rushkii/egs-watch/internal/schedule"
	"github.com/rushkii/egs-watch/internal/whatsapp"
	"github.com/rushkii/egs-watch/pkg"
)

func main() {
	http := pkg.NewClient()
	game := epic.NewEpicGames(http)

	wa, err := whatsapp.New()
	if err != nil {
		log.Fatal("Error setting up WhatsApp client:", err)
	}

	defer wa.Disconnect()

	db, err := database.New(config.PGHost, config.PGPort,
		config.PGUser, config.PGPwd, config.PGDb,
	)
	if err != nil {
		log.Println(err)
	}

	defer db.Close()

	repo := repository.NewRepository(db)

	scheduler := schedule.NewCron(repo, wa, game, http)
	scheduler.PrepareJobs()
	scheduler.Start()

	defer scheduler.Stop()

	log.Println("🤖 Bot is running and Scheduler is active. Press Ctrl+C to exit.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down client...")
}
