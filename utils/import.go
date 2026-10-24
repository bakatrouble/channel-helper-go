package utils

import (
	"channel-helper-go/database"
	"time"
)

type ImportMessage struct {
	ChatID    int64 `json:"chat_id"`
	MessageID int   `json:"message_id"`
}

type ImportItem struct {
	Type       database.MediaType `json:"type"`
	FileID     string             `json:"file_id"`
	MessageIds []int              `json:"message_ids"`
	Messages   []ImportMessage    `json:"messages,omitempty"`
	Processed  bool               `json:"processed"`
	Datetime   time.Time          `json:"datetime"`
}
