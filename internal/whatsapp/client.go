package whatsapp

import (
	"context"
	"log"
	"os"

	"github.com/mdp/qrterminal/v3"

	"go.mau.fi/whatsmeow"
)

type WhatsApp struct {
	*whatsmeow.Client
	// Game *epic.EpicGames
	// Http *pkg.HttpClient
}

func New() (*WhatsApp, error) {
	device, err := setupDevice()
	if err != nil {
		return nil, err
	}

	client := &WhatsApp{
		Client: whatsmeow.NewClient(device, nil),
		// Game:   epic.NewEpicGames(),
		// Http:   pkg.NewClient(),
	}

	client.AddEventHandler(func(evt any) {
		client.EventHandler(evt)
	})

	if client.Store.ID == nil {
		err = client.printQR()
		if err != nil {
			return nil, err
		}
	} else {
		err = client.Connect()
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (client *WhatsApp) Disconnect() {
	client.Client.Disconnect()
}

func (client *WhatsApp) printQR() error {
	qrChan, _ := client.GetQRChannel(context.Background())

	err := client.Connect()
	if err != nil {
		return err
	}

	for evt := range qrChan {
		if evt.Event == "code" {
			log.Println("Scan this QR code with WhatsApp:")
			qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
		} else {
			log.Println("Login event:", evt.Event)
		}
	}

	return nil
}
