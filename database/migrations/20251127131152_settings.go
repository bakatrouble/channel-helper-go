package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	type settings struct {
		bun.BaseModel  `bun:"table:settings,alias:s"`
		GroupThreshold int `bun:",notnull"`
	}

	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] ")
		if _, err := db.NewCreateTable().
			Model((*settings)(nil)).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		if _, err := db.NewDropTable().
			Model((*settings)(nil)).
			IfExists().
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
		return nil
	})
}
