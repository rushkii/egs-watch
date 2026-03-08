package whatsapp

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/rushkii/egs-watch/internal/config"

	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func init() {
	path := strings.TrimPrefix(config.SessionString, "file:")

	if i := strings.Index(path, "?"); i != -1 {
		path = path[:i]
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Println(err)
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			log.Println(err)
			return
		}
		f.Close()
	}

}

func setupDevice() (*store.Device, error) {
	dbLog := waLog.Stdout("Database", "INFO", true)
	ctx := context.Background()

	container, err := sqlstore.New(ctx, "sqlite3", config.SessionString, dbLog)
	if err != nil {
		return nil, err
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, err
	}

	return deviceStore, nil
}
