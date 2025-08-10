package utils

import (
	"channel-helper-go/database"
	"time"
)

type ImportItem struct {
	Type       database.MediaType `json:"type"`
	FileID     string             `json:"file_id"`
	MessageIds []int              `json:"message_ids"`
	Processed  bool               `json:"processed"`
	Datetime   time.Time          `json:"datetime"`
}
