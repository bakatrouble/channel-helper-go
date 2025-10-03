package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] ")
		if _, err := db.NewCreateIndex().
			Model((*imageHash)(nil)).
			Index("idx_image_hashes_hash_normal").
			Column("hash").
			Exec(ctx); err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		if _, err := db.NewDropIndex().
			Model((*imageHash)(nil)).
			Index("idx_image_hashes_hash_normal").
			IfExists().
			Cascade().
			Exec(ctx); err != nil {
			return err
		}
		return nil
	})
}
