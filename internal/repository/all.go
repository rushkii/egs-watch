package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
)

type Repository struct {
	*gamesRepository
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{gamesRepository: NewGamesRepository(db)}
}

func (a *ImageArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		str, okStr := value.(string)
		if !okStr {
			return errors.New("Failed to scan JSON: expected []byte or string")
		}
		bytes = []byte(str)
	}

	return json.Unmarshal(bytes, a)
}
