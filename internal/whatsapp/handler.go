package whatsapp

import (
	"log"
	"strings"

	"go.mau.fi/whatsmeow/types/events"
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
	// case "/test":
	// 	if isKizu(v) {
	// 		client.cmdTest(v)
	// 	}
	}
}

// func (client *WhatsApp) cmdTest(v *events.Message) {
// 	freeGames, err := client.Game.GetFreeGamesFromEGS()
// 	if err != nil {
// 		log.Println("Error fetching free games:", err)
// 		return
// 	}

// 	now, upcoming := epic.FormatFreeAllGames(freeGames)
// 	result := fmt.Sprintf("%s\n%s", now, upcoming)
// 	ctx := context.Background()

// 	log.Println("free games has been formatted")

// 	var messages []*waE2E.Message

// 	for index, fg := range freeGames.Now {
// 		fmt.Printf("attempt to send image %d of %d\n", index+1, len(freeGames.Now))
// 		imageUrl := epic.GetImageWide(fg.KeyImages)

// 		img, err := client.Http.Download(imageUrl)
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}

// 		fmt.Printf("downloaded image %d of %d\n", index+1, len(freeGames.Now))

// 		uploaded, err := client.Upload(context.Background(), img, whatsmeow.MediaImage)
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}

// 		mimetype := http.DetectContentType(img)
// 		fmt.Printf("uploaded image %d of %d\n", index+1, len(freeGames.Now))

// 		thumbnail, errThumb := pkg.GenerateThumbnail(img, 200)
// 		if errThumb != nil {
// 			fmt.Printf("Warning: couldn't generate thumbnail: %v\n", errThumb)
// 		}

// 		imgMsg := &waE2E.ImageMessage{
// 			Mimetype:      proto.String(mimetype),
// 			URL:           &uploaded.URL,
// 			DirectPath:    &uploaded.DirectPath,
// 			MediaKey:      uploaded.MediaKey,
// 			FileEncSHA256: uploaded.FileEncSHA256,
// 			FileSHA256:    uploaded.FileSHA256,
// 			FileLength:    &uploaded.FileLength,
// 			JPEGThumbnail: thumbnail,
// 		}

// 		if index == len(freeGames.Now)-1 {
// 			imgMsg.Caption = proto.String(result)
// 		}

// 		messages = append(messages, &waE2E.Message{
// 			ImageMessage: imgMsg,
// 		})
// 	}

// 	for index, msg := range messages {
// 		_, err = client.SendMessage(ctx, v.Info.Chat, msg)
// 		if err != nil {
// 			log.Println("Error sending message:", err)
// 		}
// 		fmt.Printf("sent message %d of %d\n", index+1, len(freeGames.Now))
// 	}
// }
