package scripts

import (
	"channel-helper-go/database"
	"channel-helper-go/utils"
	"context"
	"fmt"
	go_console "github.com/DrSmithFr/go-console"
)

func MigrateScript(cmd *go_console.Script) go_console.ExitCode {
	ctx := context.Background()

	config, err := utils.ParseConfig(cmd.Input.Option("config"))
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to parse config file: %v\n", err)
		return go_console.ExitError
	}

	logger := utils.NewLogger(config.DbName, "db")
	sqldb, err := database.NewSQLDB(config.DbName)
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to connect to database: %v\n", err)
		return go_console.ExitError
	}
	db, err := database.NewDBStruct(sqldb, true, logger)
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to create database struct: %v\n", err)
		return go_console.ExitError
	}
	defer func(db *database.DBStruct) {
		_ = db.Close()
	}(db)

	migrator := db.GetMigrator()
	if migrator == nil {
		_, _ = fmt.Fprintf(cmd, "No migrator found for database %s\n", config.DbName)
		return go_console.ExitError
	}
	if err = migrator.Lock(ctx); err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to lock database for migration: %v\n", err)
		return go_console.ExitError
	}
	defer func() {
		if err := migrator.Unlock(ctx); err != nil {
			_, _ = fmt.Fprintf(cmd, "Failed to unlock database after migration: %v\n", err)
		}
	}()

	group, err := migrator.Migrate(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to migrate database: %v\n", err)
		return go_console.ExitError
	}

	if group.IsZero() {
		_, _ = fmt.Fprintf(cmd, "Database %s is already up to date\n", config.DbName)
		return go_console.ExitSuccess
	}

	_, _ = fmt.Fprintf(cmd, "Migrated database %s to version %s\n", config.DbName, group.String())
	return go_console.ExitSuccess
}

func CreateGoMigrationScript(cmd *go_console.Script) go_console.ExitCode {
	ctx := context.Background()

	config, err := utils.ParseConfig(cmd.Input.Option("config"))
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to parse config file: %v\n", err)
		return go_console.ExitError
	}

	logger := utils.NewLogger(config.DbName, "db")
	sqldb, err := database.NewSQLDB(config.DbName)
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to connect to database: %v\n", err)
		return go_console.ExitError
	}
	db, err := database.NewDBStruct(sqldb, true, logger)
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to create database struct: %v\n", err)
		return go_console.ExitError
	}
	defer func(db *database.DBStruct) {
		_ = db.Close()
	}(db)

	migrator := db.GetMigrator()
	if migrator == nil {
		_, _ = fmt.Fprintf(cmd, "No migrator found for database %s\n", config.DbName)
		return go_console.ExitError
	}

	mf, err := migrator.CreateGoMigration(ctx, cmd.Input.Argument("name"))
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to create migration file: %v\n", err)
		return go_console.ExitError
	}

	_, _ = fmt.Fprintf(cmd, "Created migration %s (%s)\n", mf.Name, mf.Path)

	return go_console.ExitSuccess
}

func InitDBScript(cmd *go_console.Script) go_console.ExitCode {
	ctx := context.Background()

	config, err := utils.ParseConfig(cmd.Input.Option("config"))
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to parse config file: %v\n", err)
		return go_console.ExitError
	}

	logger := utils.NewLogger(config.DbName, "db")
	sqldb, err := database.NewSQLDB(config.DbName)
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to connect to database: %v\n", err)
		return go_console.ExitError
	}
	db, err := database.NewDBStruct(sqldb, true, logger)
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to create database struct: %v\n", err)
		return go_console.ExitError
	}
	defer func(db *database.DBStruct) {
		_ = db.Close()
	}(db)

	migrator := db.GetMigrator()
	if migrator == nil {
		_, _ = fmt.Fprintf(cmd, "No migrator found for database %s\n", config.DbName)
		return go_console.ExitError
	}

	if err = migrator.Init(ctx); err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to initialize database: %v\n", err)
		return go_console.ExitError
	}

	_, _ = fmt.Fprintf(cmd, "Database %s initialized successfully\n", config.DbName)
	return go_console.ExitSuccess
}
