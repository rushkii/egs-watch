package whatsapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rushkii/egs-watch/pkg"
	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func (client *WhatsApp) SendMedia(ctx context.Context, to types.JID, items ...any) error {
	imgLen, vidLen := 0, 0

	for _, item := range items {
		switch item.(type) {
		case waE2E.ImageMessage, *waE2E.ImageMessage:
			imgLen++
		case waE2E.VideoMessage, *waE2E.VideoMessage:
			vidLen++
		}
	}

	var parentKey *waCommon.MessageKey

	if imgLen+vidLen > 1 {
		albumKey, err := pkg.RandomCrypt(32)
		if err != nil {
			return fmt.Errorf("Failed to generate album key: %w", err)
		}

		albumMessage := &waE2E.Message{
			AlbumMessage: &waE2E.AlbumMessage{
				ExpectedImageCount: proto.Uint32(uint32(imgLen)),
				ExpectedVideoCount: proto.Uint32(uint32(vidLen)),
			},
			MessageContextInfo: &waE2E.MessageContextInfo{
				MessageSecret: albumKey,
			},
		}

		albumResp, err := client.SendMessage(ctx, to, albumMessage)
		if err != nil {
			return fmt.Errorf("Failed to send album container: %w", err)
		}

		parentKey = &waCommon.MessageKey{
			FromMe:    proto.Bool(true),
			ID:        proto.String(albumResp.ID),
			RemoteJID: proto.String(to.String()),
		}
	}

	for _, item := range items {
		var err error
		msg := &waE2E.Message{}

		if imgLen+vidLen > 1 {
			mediaSecret, err := pkg.RandomCrypt(32)
			if err != nil {
				continue
			}

			msg.MessageContextInfo = &waE2E.MessageContextInfo{
				MessageSecret: mediaSecret,
				MessageAssociation: &waE2E.MessageAssociation{
					AssociationType:  waE2E.MessageAssociation_MEDIA_ALBUM.Enum(),
					ParentMessageKey: parentKey,
				},
			}
		}

		switch v := item.(type) {
		case *waE2E.ImageMessage:
			msg.ImageMessage = v
		case waE2E.ImageMessage:
			msg.ImageMessage = &v
		case *waE2E.VideoMessage:
			msg.VideoMessage = v
		case waE2E.VideoMessage:
			msg.VideoMessage = &v
		default:
			continue
		}

		_, err = client.SendMessage(ctx, to, msg)
		if err != nil {
			continue
		}
	}

	return nil
}

func (client *WhatsApp) SendButton(
	ctx context.Context,
	to types.JID,
	content ButtonContent,
) error {
	nativeButtons := processButtons(content.Buttons)

	header := &waE2E.InteractiveMessage_Header{
		Title:              proto.String(content.Title),
		HasMediaAttachment: proto.Bool(false),
	}

	if content.Image != nil {
		uploaded, err := client.Upload(ctx, content.Image, whatsmeow.MediaImage)
		if err != nil {
			return fmt.Errorf("failed to upload button image: %w", err)
		}

		mimetype := http.DetectContentType(content.Image)
		thumbnail, _ := pkg.GenerateThumbnail(content.Image, 200)

		header.HasMediaAttachment = proto.Bool(true)
		header.Media = &waE2E.InteractiveMessage_Header_ImageMessage{
			ImageMessage: &waE2E.ImageMessage{
				Mimetype:      proto.String(mimetype),
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				JPEGThumbnail: thumbnail,
			},
		}
	} else if content.Video != nil {
		uploaded, err := client.Upload(ctx, content.Video, whatsmeow.MediaVideo)
		if err != nil {
			return fmt.Errorf("failed to upload button video: %w", err)
		}

		mimetype := http.DetectContentType(content.Video)
		thumbnail, _ := pkg.GenerateThumbnail(content.Video, 200)

		header.HasMediaAttachment = proto.Bool(true)
		header.Media = &waE2E.InteractiveMessage_Header_VideoMessage{
			VideoMessage: &waE2E.VideoMessage{
				Mimetype:      proto.String(mimetype),
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				JPEGThumbnail: thumbnail,
			},
		}
	}

	interactiveMsg := &waE2E.InteractiveMessage{
		Header: header,
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String(content.Text),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String(content.Footer),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons: nativeButtons,
			},
		},
	}

	msg := &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				InteractiveMessage: interactiveMsg,
			},
		},
	}

	additionalNodes := []waBinary.Node{
		{
			Tag:   "biz",
			Attrs: waBinary.Attrs{},
			Content: []waBinary.Node{
				{
					Tag:   "interactive",
					Attrs: waBinary.Attrs{"type": "native_flow", "v": "1"},
					Content: []waBinary.Node{
						{
							Tag:   "native_flow",
							Attrs: waBinary.Attrs{"v": "9", "name": "mixed"},
						},
					},
				},
			},
		},
	}

	_, err := client.SendMessage(ctx, to, msg,
		whatsmeow.SendRequestExtra{
			AdditionalNodes: &additionalNodes,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send button message: %w", err)
	}

	return nil
}
