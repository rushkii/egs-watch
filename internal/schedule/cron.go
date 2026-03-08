package schedule

import (
	"github.com/robfig/cron/v3"
	"github.com/rushkii/egs-watch/internal/epic"
	"github.com/rushkii/egs-watch/internal/repository"
	"github.com/rushkii/egs-watch/internal/whatsapp"
	"github.com/rushkii/egs-watch/pkg"
)

type Scheduler struct {
	Cron     *cron.Cron
	Repo     *repository.Repository
	WhatsApp *whatsapp.WhatsApp
	Game     *epic.EpicGames
	Http     *pkg.HttpClient
}

func NewCron(repo *repository.Repository, wa *whatsapp.WhatsApp, game *epic.EpicGames, hc *pkg.HttpClient) *Scheduler {
	return &Scheduler{
		Cron:     cron.New(),
		Repo:     repo,
		WhatsApp: wa,
		Game:     game,
		Http:     hc,
	}
}

func (s *Scheduler) Start() {
	s.Cron.Start()
	s.TriggerSendFreeGamesUpdate()
}

func (s *Scheduler) Stop() {
	s.Cron.Stop()
}
