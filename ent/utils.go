package ent

import (
	"channel-helper-go/ent/imagehash"
	"context"
	"entgo.io/ent/dialect"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"log/slog"
	"path"
)

func ConnectToDB(dbName string, ctx context.Context) (*Client, error) {
	client, err := Open(
		dialect.SQLite,
		fmt.Sprintf("file:%s?cache=shared&_fk=1", path.Join("dbs", fmt.Sprintf("%s.sqlite", dbName))),
	)
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
		return nil, err
	}
	//err = client.Schema.Create(ctx)
	//if err != nil {
	//	log.Fatalf("failed creating schema resources: %v", err)
	//}
	return client, nil
}

func ImageHashExists(hash string, ctx context.Context, client *Client, logger *slog.Logger) (bool, *Post, *UploadTask, error) {
	imageHash, err := client.ImageHash.Query().
		WithPost().
		WithUploadTask().
		Where(imagehash.ImageHashEQ(hash)).
		First(ctx)
	if imageHash != nil {
		return true, imageHash.Edges.Post, imageHash.Edges.UploadTask, nil
	} else if err != nil && !IsNotFound(err) {
		logger.With("err", err).Error("failed to check image hash existence")
		return false, nil, nil, err
	}
	return false, nil, nil, nil
}
