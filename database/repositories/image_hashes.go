package repositories

import (
	"channel-helper-go/database/schema"
	"context"
	"database/sql"
	"errors"
	"github.com/uptrace/bun"
)

type ImageHashRepository struct {
	db *bun.DB
}

func NewImageHashRepository(db *bun.DB) *ImageHashRepository {
	return &ImageHashRepository{
		db: db,
	}
}

func (r *ImageHashRepository) Exists(ctx context.Context, hash string) (bool, *schema.Post, *schema.UploadTask, error) {
	imageHash := new(schema.ImageHash)
	if err := r.db.NewSelect().
		Model(imageHash).
		Where("hash = ?", hash).
		Relation("Post").
		Relation("UploadTask").
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil, nil, nil // Image hash does not exist
		}
		return false, nil, nil, err // Other error occurred
	}
	return true, imageHash.Post, imageHash.UploadTask, nil
}

func (r *ImageHashRepository) GetAll(ctx context.Context) ([]string, error) {
	imageHashes := make([]string, 0)
	if err := r.db.NewSelect().
		Model(&schema.ImageHash{}).
		Column("hash").
		Scan(ctx, &imageHashes); err != nil {
		return nil, err
	}
	return imageHashes, nil
}
