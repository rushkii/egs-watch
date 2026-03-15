package schedule

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

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
		log.Printf("Error parsing JID: %v", err)
		return
	}

	freeGames, err := s.Repo.GetFreeGamesFromDB()
	if err != nil {
		log.Printf("Error fetching free games: %v", err)
		return
	}

	if len(freeGames.Now) == 0 || freeGames.Now == nil {
		log.Println("Free games update already sent")
		return
	}

	now, upcoming := repository.FormatFreeAllGames(freeGames)
	result := fmt.Sprintf("%s\n%s", now, upcoming)
	ctx := context.Background()

	items := make([]*waE2E.ImageMessage, len(freeGames.Now))
	fgids := make([]string, len(freeGames.Now))
	var wg sync.WaitGroup

	for i, fg := range freeGames.Now {
		fgids[i] = fg.ID
		wg.Add(1)
		go s.processGameImage(&wg, i, ctx, fg, items)
	}

	wg.Wait()

	var albumItems []any
	var validFgids []string
	captionSet := false

	for i, item := range items {
		if item == nil {
			continue
		}

		if !captionSet {
			item.Caption = proto.String(result)
			captionSet = true
		}

		albumItems = append(albumItems, item)
		validFgids = append(validFgids, fgids[i])
	}

	if len(albumItems) != 0 {
		err = s.WhatsApp.SendMedia(ctx, target, albumItems...)
		if err != nil {
			log.Printf("Failed to send media: %v", err)
		}

		for _, fgid := range validFgids {
			if err := s.Repo.InsertUpdateSent(fgid); err != nil {
				log.Printf("Error inserting the free games update sent: %v\n", err)
			}
		}
	}

	log.Println("Cron triggered: Free Games update has been sent!")
}

func (s *Scheduler) processGameImage(
	wg *sync.WaitGroup, index int,
	ctx context.Context,
	game repository.FreeGamesFromDB,
	items []*waE2E.ImageMessage,
) {
	defer wg.Done()

	log.Printf("Processing image %d\n", index+1)
	imageUrl := repository.GetImageWide(game.Images)

	img, err := s.Http.Download(imageUrl)
	if err != nil {
		log.Printf("Error downloading image %d: %v\n", index+1, err)
		return
	}

	uploaded, err := s.WhatsApp.Upload(ctx, img, whatsmeow.MediaImage)
	if err != nil {
		log.Printf("Error uploading image %d: %v\n", index+1, err)
		return
	}

	mimetype := http.DetectContentType(img)
	thumbnail, errThumb := pkg.GenerateThumbnail(img, 200)
	if errThumb != nil {
		log.Printf("Warning: couldn't generate thumbnail for %d: %v\n", index+1, errThumb)
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

	items[index] = imgMsg
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
		log.Printf("Error cleaning up stale games: %v\n", err)
	}

	log.Printf("Cleanup complete. Deleted %d stale games.\n", result)
}
