package schema

import (
	"channel-helper-go/database/database_utils"
	"context"
	"time"

	"github.com/uptrace/bun"
)

type MediaType string

const (
	MediaTypePhoto     MediaType = "photo"
	MediaTypeVideo     MediaType = "video"
	MediaTypeAnimation MediaType = "animation"
)

type ImageHash struct {
	bun.BaseModel `bun:"table:image_hashes,alias:ih"`
	ID            int64       `bun:",pk,autoincrement"`
	Hash          string      `bun:",unique,notnull"`
	Post          *Post       `bun:"rel:has-one,join:id=image_hash_id"`
	UploadTask    *UploadTask `bun:"rel:has-one,join:id=image_hash_id"`
}

type Post struct {
	bun.BaseModel `bun:"table:posts,alias:p"`
	ID            string    `bun:",pk"`
	Type          MediaType `bun:",notnull"`
	FileID        string
	IsSent        bool      `bun:",default:false"`
	CreatedAt     time.Time `bun:",default:current_timestamp"`
	SentAt        *time.Time
	ImageHashID   *int64
	ImageHash     *ImageHash   `bun:"rel:belongs-to,join:image_hash_id=id"`
	MessageIDs    []*MessageID `bun:"rel:has-many,join:id=post_id"`
}

type UploadTask struct {
	bun.BaseModel `bun:"table:upload_tasks,alias:ut"`
	ID            string    `bun:",pk"`
	Type          MediaType `bun:",notnull"`
	Data          *[]byte
	IsProcessed   bool      `bun:"default:false"`
	CreatedAt     time.Time `bun:"default:current_timestamp"`
	SentAt        *time.Time
	ImageHashID   *int64
	ImageHash     *ImageHash `bun:"rel:belongs-to,join:image_hash_id=id"`
}

type MessageID struct {
	bun.BaseModel `bun:"table:message_ids,alias:mi"`
	ChatID        int64
	MessageID     int
	PostID        string
	Post          Post `bun:"rel:belongs-to,join:post_id=id"`
}

type Settings struct {
	bun.BaseModel  `bun:"table:settings,alias:s"`
	GroupThreshold int `bun:",notnull"`
}

func (p *Post) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if p.ID == "" {
			p.ID = database_utils.GenerateID()
		}
	}
	return nil
}

func (p *UploadTask) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if p.ID == "" {
			p.ID = database_utils.GenerateID()
		}
	}
	return nil
}
