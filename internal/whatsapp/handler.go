package whatsapp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/rushkii/egs-watch/internal/epic"
	"github.com/rushkii/egs-watch/pkg"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func (client *WhatsApp) EventHandler(evt any) {
	switch v := evt.(type) {

	case *events.Connected:
		log.Println("✅ WhatsApp connected!")

	case *events.Disconnected:
		log.Println("❌ WhatsApp disconnected:", v)

	case *events.Message:
		text := v.Message.GetConversation()
		if text == "" && v.Message.GetExtendedTextMessage() != nil {
			text = v.Message.GetExtendedTextMessage().GetText()
		}

		if strings.HasPrefix(text, "/") && !v.Info.IsFromMe {
			client.handleCommand(v, text)
		}

	default:

	}
}

func (client *WhatsApp) handleCommand(v *events.Message, text string) {
	switch strings.TrimSpace(text) {
	case "/test":
		if isKizu(v) {
			client.cmdTest(v)
		}
	}
}

func (client *WhatsApp) cmdTest(v *events.Message) {
	network := pkg.NewClient()
	game := epic.NewEpicGames(network)

	freeGames, err := game.GetFreeGamesFromEGS()
	if err != nil {
		log.Println("Error fetching free games:", err)
		return
	}

	now, upcoming := epic.FormatFreeAllGames(freeGames)
	result := fmt.Sprintf("%s\n%s", now, upcoming)

	ctx := context.Background()
	chat := v.Info.Chat

	items := make([]*waE2E.ImageMessage, len(freeGames.Now))
	var wg sync.WaitGroup

	for i, fg := range freeGames.Now {
		wg.Add(1)
		go client.processGameImage(&wg, i, ctx, network, fg, items)
	}

	wg.Wait()

	var albumItems []any
	captionSet := false

	for _, item := range items {
		if item == nil {
			continue
		}

		if !captionSet {
			item.Caption = proto.String(result)
			captionSet = true
		}

		albumItems = append(albumItems, item)
	}

	if len(albumItems) != 0 {
		err = client.SendMedia(ctx, chat, albumItems...)
		if err != nil {
			log.Printf("Failed to send album: %v", err)
		}
	}
}

func (client *WhatsApp) processGameImage(
	wg *sync.WaitGroup, index int,
	ctx context.Context,
	network *pkg.HttpClient,
	game epic.FGElement,
	items []*waE2E.ImageMessage,
) {
	defer wg.Done()

	imageUrl := epic.GetImageWide(game.KeyImages)
	img, err := network.Download(imageUrl)
	if err != nil {
		log.Printf("Error downloading image %d: %v\n", index+1, err)
		return
	}

	uploaded, err := client.Upload(ctx, img, whatsmeow.MediaImage)
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
