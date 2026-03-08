package whatsapp

import "go.mau.fi/whatsmeow/types/events"

func isKizu(v *events.Message) bool {
	return v.Info.Sender.User == "6281292942010" || v.Info.Sender.User == "32783602810885"
}
