package repositories

import "github.com/uptrace/bun"

func rollbackIfError(err error, tx bun.Tx) error {
	if err == nil {
		return nil
	}
	if errRollback := tx.Rollback(); errRollback != nil {
		return errRollback
	}
	return err
}
