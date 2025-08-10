package repositories

import (
	"channel-helper-go/database/schema"
	"context"
	"database/sql"
	"errors"
	"github.com/uptrace/bun"
)

type PostRepository struct {
	db *bun.DB
}

func NewPostRepository(db *bun.DB) *PostRepository {
	return &PostRepository{
		db: db,
	}
}

func (r *PostRepository) CreateBulk(ctx context.Context, posts []*schema.Post, chunkSize int) error {
	if len(posts) == 0 {
		return nil
	}

	var tx bun.Tx
	var err error
	if tx, err = r.db.BeginTx(ctx, nil); err != nil {
		return err
	}

	for i := 0; i < len(posts); i += chunkSize {
		end := i + chunkSize
		if end > len(posts) {
			end = len(posts)
		}
		chunk := posts[i:end]

		for _, post := range chunk {
			if post.ImageHash != nil && post.ImageHash.ID == 0 {
				if _, err = tx.NewInsert().Model(post.ImageHash).Exec(ctx); err != nil {
					return rollbackIfError(err, tx)
				}
				post.ImageHashID = &post.ImageHash.ID
			}

			if _, err = tx.NewInsert().Model(post).Exec(ctx); err != nil {
				return rollbackIfError(err, tx)
			}

			for _, messageID := range post.MessageIDs {
				if messageID.PostID == "" {
					messageID.PostID = post.ID
					if _, err = tx.NewInsert().Model(&messageID).Exec(ctx); err != nil {
						return rollbackIfError(err, tx)
					}
				}
			}
		}
	}

	return tx.Commit()
}

func (r *PostRepository) Create(ctx context.Context, post *schema.Post) error {
	if post == nil {
		return nil
	}

	var tx bun.Tx
	var err error
	if tx, err = r.db.BeginTx(ctx, nil); err != nil {
		return err
	}

	if post.ImageHash != nil {
		if post.ImageHash.ID == 0 {
			if _, err = tx.NewInsert().Model(post.ImageHash).Exec(ctx); err != nil {
				return rollbackIfError(err, tx)
			}
		}
		post.ImageHashID = &post.ImageHash.ID
	}

	if _, err = tx.NewInsert().Model(post).Exec(ctx); err != nil {
		return rollbackIfError(err, tx)
	}

	for _, messageID := range post.MessageIDs {
		if messageID.PostID == "" {
			messageID.PostID = post.ID
			if _, err = tx.NewInsert().Model(&messageID).Exec(ctx); err != nil {
				return rollbackIfError(err, tx)
			}
		}
	}

	return tx.Commit()
}

func (r *PostRepository) Update(ctx context.Context, post *schema.Post) error {
	if post == nil {
		return errors.New("post cannot be nil")
	}
	if post.ID == "" {
		return errors.New("post ID cannot be empty")
	}

	var tx bun.Tx
	var err error
	if tx, err = r.db.BeginTx(ctx, nil); err != nil {
		return err
	}

	if _, err = tx.NewUpdate().Model(post).WherePK().Exec(ctx); err != nil {
		return rollbackIfError(err, tx)
	}

	for _, messageID := range post.MessageIDs {
		if messageID.PostID == "" {
			messageID.PostID = post.ID
			if _, err = tx.NewInsert().Model(&messageID).Exec(ctx); err != nil {
				return rollbackIfError(err, tx)
			}
		}
	}

	if post.ImageHash != nil {
		if post.ImageHash.ID == 0 {
			if _, err = tx.NewInsert().Model(post.ImageHash).Exec(ctx); err != nil {
				return rollbackIfError(err, tx)
			}
		}
		post.ImageHashID = &post.ImageHash.ID
	}

	return tx.Commit()
}

func (r *PostRepository) UnsentCount(ctx context.Context) (int, error) {
	var count int

	if _, err := r.db.NewSelect().
		Model((*schema.Post)(nil)).
		Where("is_sent = ?", false).
		Count(ctx); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *PostRepository) GetByMessageID(ctx context.Context, chatID int64, messageID int) (*schema.Post, error) {
	var post schema.Post
	if err := r.db.NewSelect().
		Model(&post).
		Relation("MessageIDs").
		Relation("ImageHash").
		Where("message_ids.chat_id = ? AND message_ids.message_id = ?", chatID, messageID).
		Scan(ctx); err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) Delete(ctx context.Context, post *schema.Post) error {
	if post == nil {
		return errors.New("post cannot be nil")
	}
	if post.ID == "" {
		return errors.New("post ID cannot be empty")
	}

	var tx bun.Tx
	var err error
	if tx, err = r.db.BeginTx(ctx, nil); err != nil {
		return err
	}

	if _, err = tx.NewDelete().Model(post).WherePK().Exec(ctx); err != nil {
		return rollbackIfError(err, tx)
	}

	if post.ImageHash != nil && post.ImageHash.ID != 0 {
		if _, err = tx.NewDelete().Model(post.ImageHash).WherePK().Exec(ctx); err != nil {
			return rollbackIfError(err, tx)
		}
	}

	if _, err = tx.NewDelete().Model(post.MessageIDs).WherePK().Exec(ctx); err != nil {
		return rollbackIfError(err, tx)
	}

	return tx.Commit()
}

func (r *PostRepository) GetAll(ctx context.Context) ([]*schema.Post, error) {
	var posts []*schema.Post
	if err := r.db.NewSelect().
		Model(&posts).
		Relation("MessageIDs").
		Relation("ImageHash").
		Scan(ctx); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) GetRandomUnsent(ctx context.Context) (*schema.Post, error) {
	var post schema.Post
	if err := r.db.NewSelect().
		Model(&post).
		Relation("MessageIDs").
		Relation("ImageHash").
		Where("is_sent = ?", false).
		OrderExpr("random()").
		Limit(1).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) GetFileIDs(ctx context.Context) (map[string]bool, error) {
	fileIds := make([]string, 0)
	if err := r.db.NewSelect().
		Model((*schema.Post)(nil)).
		Column("file_id").
		Scan(ctx, &fileIds); err != nil {
		return nil, err
	}
	fileIdSet := make(map[string]bool, len(fileIds))
	for _, fileId := range fileIds {
		fileIdSet[fileId] = true
	}
	return fileIdSet, nil
}
