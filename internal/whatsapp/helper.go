package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/rushkii/egs-watch/internal/epic"
	"github.com/rushkii/egs-watch/pkg"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func isKizu(v *events.Message) bool {
	return v.Info.Sender.User == "6281292942010" || v.Info.Sender.User == "32783602810885"
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

func processButtons(buttons []*waE2E.ButtonsMessage_Button) []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton {
	processed := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0, len(buttons))

	for i, btn := range buttons {
		if btn.NativeFlowInfo != nil &&
			btn.NativeFlowInfo.Name != nil &&
			btn.NativeFlowInfo.ParamsJSON != nil {
			processed = append(processed, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
				Name:             btn.NativeFlowInfo.Name,
				ButtonParamsJSON: btn.NativeFlowInfo.ParamsJSON,
			})
			continue
		}

		displayText := fmt.Sprintf("Button %d", i+1)
		if btn.ButtonText.GetDisplayText() != "" {
			displayText = btn.ButtonText.GetDisplayText()
		}

		id := fmt.Sprintf("id_%d", i+1)
		if btn.GetButtonID() != "" {
			id = btn.GetButtonID()
		}

		paramsJSON, _ := json.Marshal(map[string]string{
			"display_text": displayText,
			"id":           id,
		})

		processed = append(processed, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String("quick_reply"),
			ButtonParamsJSON: proto.String(string(paramsJSON)),
		})
	}

	return processed
}
