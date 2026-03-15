package pkg

import (
	"crypto/rand"
	"log"
)

func RandomCrypt(length int) ([]byte, error) {
	key := make([]byte, length)

	_, err := rand.Read(key)
	if err != nil {
		log.Fatalf("error reading random bytes: %v", err)
		return nil, err
	}

	return key, nil
}
