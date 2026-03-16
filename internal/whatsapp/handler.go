package whatsapp

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/rushkii/egs-watch/internal/epic"
	"github.com/rushkii/egs-watch/pkg"
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
	case "/freegames":
		if isKizu(v) {
			client.cmdFreeGamesTest(v)
		}
	case "/buttons":
		if isKizu(v) {
			client.cmdButtonsTest(v)
		}
	}
}

func (client *WhatsApp) cmdFreeGamesTest(v *events.Message) {
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

func (client *WhatsApp) cmdButtonsTest(v *events.Message) {
	ctx := context.Background()

	img, err := os.ReadFile("storage/pengmin.jpg")
	if err != nil {
		log.Printf("Failed to read image: %v", err)
		return
	}

	content := ButtonContent{
		Title:  "Button Title",
		Text:   "Button Content",
		Footer: "Button Footer",
		Image:  img,
		Buttons: []*waE2E.ButtonsMessage_Button{
			{
				ButtonID: proto.String("btn_1"),
				ButtonText: &waE2E.ButtonsMessage_Button_ButtonText{
					DisplayText: proto.String("Click Me!"),
				},
			},
		},
	}

	if err := client.SendButton(ctx, v.Info.Chat, content); err != nil {
		log.Printf("Failed to send button: %v", err)
	}
}
