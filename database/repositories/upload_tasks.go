package repositories

import (
	"channel-helper-go/database/schema"
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"
)

type UploadTaskRepository struct {
	db *bun.DB
}

func NewUploadTaskRepository(db *bun.DB) *UploadTaskRepository {
	return &UploadTaskRepository{
		db: db,
	}
}

func (r *UploadTaskRepository) Create(ctx context.Context, task *schema.UploadTask) error {
	if task == nil {
		return nil
	}

	var tx bun.Tx
	var err error
	if tx, err = r.db.BeginTx(ctx, nil); err != nil {
		return err
	}

	if task.ImageHash != nil {
		if task.ImageHash.ID == 0 {
			if _, err = tx.NewInsert().Model(task.ImageHash).Exec(ctx); err != nil {
				return rollbackIfError(err, tx)
			}
		}
		task.ImageHashID = &task.ImageHash.ID
	}

	if _, err = tx.NewInsert().Model(task).Exec(ctx); err != nil {
		return rollbackIfError(err, tx)
	}

	return tx.Commit()
}

func (r *UploadTaskRepository) Update(ctx context.Context, task *schema.UploadTask) error {
	if task == nil {
		return nil
	}
	if task.ID == "" {
		return errors.New("task ID cannot be empty")
	}

	var tx bun.Tx
	var err error
	if tx, err = r.db.BeginTx(ctx, nil); err != nil {
		return err
	}

	if task.ImageHash != nil {
		if task.ImageHash.ID == 0 {
			if _, err = tx.NewInsert().Model(task.ImageHash).Exec(ctx); err != nil {
				return rollbackIfError(err, tx)
			}
		}
		task.ImageHashID = &task.ImageHash.ID
	}

	if _, err = r.db.NewUpdate().Model(task).Where("id = ?", task.ID).Exec(ctx); err != nil {
		return rollbackIfError(err, tx)
	}

	return tx.Commit()
}

func (r *UploadTaskRepository) GetUnsent(ctx context.Context) ([]*schema.UploadTask, error) {
	var tasks []*schema.UploadTask
	if err := r.db.NewSelect().
		Model(&tasks).
		Where("is_processed = false").
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No unsent tasks found
		}
		return nil, err
	}
	return tasks, nil
}

func (r *UploadTaskRepository) GetByID(ctx context.Context, id string) (*schema.UploadTask, error) {
	task := new(schema.UploadTask)
	if err := r.db.NewSelect().
		Model(task).
		WherePK("id = ?", id).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No unsent tasks found
		}
		return nil, err
	}
	return task, nil
}
