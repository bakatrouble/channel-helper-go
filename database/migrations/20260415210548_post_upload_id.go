package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	type post struct {
		bun.BaseModel `bun:"table:posts,alias:p"`
		UploadTaskID  *string
	}

	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] ")
		if _, err := db.NewAddColumn().
			Model((*post)(nil)).
			ColumnExpr("upload_task_id TEXT NULL").
			Exec(ctx); err != nil {
			return err
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		if _, err := db.NewDropColumn().
			Model((*post)(nil)).
			ColumnExpr("upload_task_id").
			Exec(ctx); err != nil {
			return err
		}
		return nil
	})
}
