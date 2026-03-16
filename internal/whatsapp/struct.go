package whatsapp

import "go.mau.fi/whatsmeow/proto/waE2E"

type ButtonContent struct {
	Text    string
	Footer  string
	Title   string
	Buttons []*waE2E.ButtonsMessage_Button
	Image   []byte
	Video   []byte
}
