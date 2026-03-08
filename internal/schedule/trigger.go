package schedule

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/rushkii/egs-watch/internal/config"
	"github.com/rushkii/egs-watch/internal/repository"
	"github.com/rushkii/egs-watch/pkg"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func (s *Scheduler) TriggerSendFreeGamesUpdate() {
	log.Println("Cron triggered: Preparing to send Free Games update...")

	target, err := types.ParseJID(config.TargetGroupId)
	if err != nil {
		log.Println("Error parsing JID:", err)
		return
	}

	freeGames, err := s.Repo.GetFreeGamesFromDB()
	if err != nil {
		log.Println("Error fetching free games:", err)
		return
	}

	if len(freeGames.Now) == 0 || freeGames.Now == nil {
		log.Println("Free games update already sent")
		return
	}

	now, upcoming := repository.FormatFreeAllGames(freeGames)
	result := fmt.Sprintf("%s\n%s", now, upcoming)
	ctx := context.Background()

	var messages []*waE2E.Message

	for index, fg := range freeGames.Now {
		imageUrl := repository.GetImageWide(fg.Images)

		img, err := s.Http.Download(imageUrl)
		if err != nil {
			log.Println(err)
			continue
		}

		uploaded, err := s.WhatsApp.Upload(context.Background(), img, whatsmeow.MediaImage)
		if err != nil {
			log.Println(err)
			continue
		}

		mimetype := http.DetectContentType(img)

		thumbnail, errThumb := pkg.GenerateThumbnail(img, 200)
		if errThumb != nil {
			fmt.Printf("Warning: couldn't generate thumbnail: %v\n", errThumb)
		}

		imgMsg := &waE2E.ImageMessage{
			Mimetype:      proto.String(mimetype),
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			JPEGThumbnail: thumbnail,
		}

		if index == len(freeGames.Now)-1 {
			imgMsg.Caption = proto.String(result)
		}

		messages = append(messages, &waE2E.Message{
			ImageMessage: imgMsg,
		})

		if err := s.Repo.InsertUpdateSent(fg.ID); err != nil {
			log.Println("Error inserting the free games update sent:", err)
		}
	}

	for _, msg := range messages {
		_, err = s.WhatsApp.SendMessage(ctx, target, msg)
		if err != nil {
			log.Println("Error sending message:", err)
		}
	}

	log.Println("Cron triggered: Free Games update has been sent!")
}

func (s *Scheduler) TriggerCrawlFreeGamesData() error {
	freeGames, err := s.Game.GetFreeGamesFromEGS()
	if err != nil {
		return err
	}

	if err := s.Repo.InsertFreeGames(freeGames.All); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) TriggerCleanup() {
	log.Println("Running automatic database cleanup...")

	result, err := s.Repo.CleanupFreeGames()
	if err != nil {
		log.Println("Error cleaning up stale games: %v\n", err)
	}

	log.Printf("Cleanup complete. Deleted %d stale games.\n", result)
}
