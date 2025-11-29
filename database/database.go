package database

import (
	"channel-helper-go/database/migrations"
	"channel-helper-go/database/repositories"
	"channel-helper-go/database/schema"
	"database/sql"
	"fmt"
	"log"
	"log/slog"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/migrate"
)

type Post = schema.Post
type UploadTask = schema.UploadTask
type ImageHash = schema.ImageHash
type MessageID = schema.MessageID
type MediaType = schema.MediaType

const MediaTypeAnimation = schema.MediaTypeAnimation
const MediaTypePhoto = schema.MediaTypePhoto
const MediaTypeVideo = schema.MediaTypeVideo

type DBStruct struct {
	db         *bun.DB
	Post       *repositories.PostRepository
	UploadTask *repositories.UploadTaskRepository
	ImageHash  *repositories.ImageHashRepository
	Settings   *repositories.SettingsRepository
}

func NewSQLDB(dbName string) (*sql.DB, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, fmt.Sprintf("file:dbs/%s.sqlite3?cache=shared&mode=rwc", dbName))
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
		return nil, err
	}
	return sqldb, nil
}

func NewDBStruct(sqldb *sql.DB, verbose bool, logger *slog.Logger) (*DBStruct, error) {
	db := bun.NewDB(sqldb, sqlitedialect.New())
	queryLogLevel := slog.Level(-5) // -5 is lower than Debug level
	if verbose {
		queryLogLevel = slog.LevelDebug // Set to Debug level if verbose
	}
	db.AddQueryHook(newLogQueryHook(
		withLogger(logger),
		withQueryLogLevel(queryLogLevel),
	))

	return &DBStruct{
		db:         db,
		Post:       repositories.NewPostRepository(db),
		UploadTask: repositories.NewUploadTaskRepository(db),
		ImageHash:  repositories.NewImageHashRepository(db),
		Settings:   repositories.NewSettingsRepository(db),
	}, nil
}

func (dbs *DBStruct) GetMigrator() *migrate.Migrator {
	if dbs.db == nil {
		return nil
	}
	return migrate.NewMigrator(dbs.db, migrations.Migrations)
}

func (dbs *DBStruct) Close() error {
	if dbs.db == nil {
		return nil
	}
	if err := dbs.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	dbs.db = nil
	return nil
}
