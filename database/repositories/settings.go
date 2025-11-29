package repositories

import (
	"channel-helper-go/database/schema"
	"channel-helper-go/utils"
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"
)

type SettingsRepository struct {
	db *bun.DB
}

func NewSettingsRepository(db *bun.DB) *SettingsRepository {
	return &SettingsRepository{
		db,
	}
}

func (r *SettingsRepository) Get(ctx context.Context, config *utils.Config) (*schema.Settings, error) {
	var settings schema.Settings
	if err := r.db.NewSelect().
		Model(&settings).
		Where("1 = 1").
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			settings = schema.Settings{
				GroupThreshold: config.GroupThreshold,
			}
		} else {
			return nil, err
		}
	}
	return &settings, nil
}

func (r *SettingsRepository) Save(ctx context.Context, settings *schema.Settings) error {
	var tx bun.Tx
	var err error
	if tx, err = r.db.BeginTx(ctx, nil); err != nil {
		return err
	}
	if _, err := tx.NewDelete().
		Model((*schema.Settings)(nil)).
		Where("1 = 1").
		Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.NewInsert().
		Model(settings).
		Exec(ctx); err != nil {
		return err
	}
	return tx.Commit()
}
