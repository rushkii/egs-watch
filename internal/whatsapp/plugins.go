package whatsapp

import (
	"context"
	"fmt"

	"github.com/rushkii/egs-watch/pkg"
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
