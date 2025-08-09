package utils

import (
	"channel-helper-go/ent/post"
	"time"
)

type ImportItem struct {
	Type       post.Type `json:"type"`
	FileId     string    `json:"file_id"`
	MessageIds []int     `json:"message_ids"`
	Processed  bool      `json:"processed"`
	Datetime   time.Time `json:"datetime"`
}
