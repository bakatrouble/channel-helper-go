package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type mediaType string

type imageHash struct {
	bun.BaseModel `bun:"table:image_hashes,alias:ih"`
	ID            int64       `bun:",pk,autoincrement"`
	Hash          string      `bun:",unique,notnull"`
	Post          *post       `bun:"rel:has-one,join:id=image_hash_id"`
	UploadTask    *uploadTask `bun:"rel:has-one,join:id=image_hash_id"`
}

type post struct {
	bun.BaseModel `bun:"table:posts,alias:p"`
	ID            string    `bun:",pk"`
	Type          mediaType `bun:",notnull"`
	FileID        string
	IsSent        bool      `bun:",default:false"`
	CreatedAt     time.Time `bun:",default:current_timestamp"`
	SentAt        *time.Time
	ImageHashID   *int64
	ImageHash     *imageHash  `bun:"rel:belongs-to,join:image_hash_id=id"`
	MessageIDs    []messageId `bun:"rel:has-many,join:id=post_id"`
}

type uploadTask struct {
	bun.BaseModel `bun:"table:upload_tasks,alias:ut"`
	ID            string    `bun:",pk"`
	Type          mediaType `bun:",notnull"`
	Data          *[]byte
	IsProcessed   bool      `bun:"default:false"`
	CreatedAt     time.Time `bun:"default:current_timestamp"`
	SentAt        *time.Time
	ImageHashID   *int64
	ImageHash     *imageHash `bun:"rel:belongs-to,join:image_hash_id=id"`
}

type messageId struct {
	bun.BaseModel `bun:"table:message_ids,alias:mi"`
	ChatID        int64
	MessageID     int
	PostID        string
	Post          post `bun:"rel:belongs-to,join:post_id=id"`
}

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] ")

		// ImageHash table

		_, err := db.NewCreateTable().
			Model((*imageHash)(nil)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create image_hashes table: %w", err)
		}

		_, err = db.NewCreateIndex().
			Model((*imageHash)(nil)).
			Index("idx_image_hashes_hash").
			Column("hash").
			Unique().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create index on image_hashes: %w", err)
		}

		// Post table

		_, err = db.NewCreateTable().
			Model((*post)(nil)).
			WithForeignKeys().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create posts table: %w", err)
		}

		_, err = db.NewCreateIndex().
			Model((*post)(nil)).
			Index("idx_posts_issent").
			Column("is_sent").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create index on posts table: %w", err)
		}

		_, err = db.NewCreateIndex().
			Model((*post)(nil)).
			Index("idx_posts_type").
			Column("type").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create index on posts table: %w", err)
		}

		// UploadTask table

		_, err = db.NewCreateTable().
			Model((*uploadTask)(nil)).
			WithForeignKeys().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create upload_tasks table: %w", err)
		}

		_, err = db.NewCreateIndex().
			Model((*uploadTask)(nil)).
			Index("idx_upload_tasks_isprocessed").
			Column("is_processed").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create index on upload_tasks table: %w", err)
		}

		// MessageId table

		_, err = db.NewCreateTable().
			Model((*messageId)(nil)).
			WithForeignKeys().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create message_ids table: %w", err)
		}

		_, err = db.NewCreateIndex().
			Model((*messageId)(nil)).
			Index("idx_message_ids_chatid_messageid").
			Column("chat_id", "message_id").
			Unique().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create index on message_ids table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")

		// MessageId table

		_, err := db.NewDropIndex().
			Model((*messageId)(nil)).
			Index("idx_message_ids_chatid_messageid").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to drop index on message_ids table: %w", err)
		}

		_, err = db.NewDropTable().
			Model((*messageId)(nil)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to drop message_ids table: %w", err)
		}

		// UploadTask table

		_, err = db.NewDropIndex().
			Model((*uploadTask)(nil)).
			Index("idx_upload_tasks_isprocessed").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to drop index on upload_tasks table: %w", err)
		}

		_, err = db.NewDropTable().
			Model((*uploadTask)(nil)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to drop upload_tasks table: %w", err)
		}

		// Post table

		_, err = db.NewDropIndex().
			Model((*post)(nil)).
			Index("idx_posts_type").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to drop index on posts table: %w", err)
		}

		_, err = db.NewDropIndex().
			Model((*post)(nil)).
			Index("idx_posts_issent").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to drop index on posts table: %w", err)
		}

		_, err = db.NewDropTable().
			Model((*post)(nil)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to drop posts table: %w", err)
		}

		// ImageHash table

		_, err = db.NewDropIndex().
			Model((*imageHash)(nil)).
			Index("idx_image_hashes_hash").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to drop index on image_hashes table: %w", err)
		}

		_, err = db.NewDropTable().
			Model((*imageHash)(nil)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to drop image_hashes table: %w", err)
		}

		return nil
	})
}
