package ent

import (
	"channel-helper-go/ent/post"
	"channel-helper-go/ent/uploadtask"
	"context"
	"entgo.io/ent/dialect"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
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
	err = client.Schema.Create(ctx)
	if err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	return client, nil
}

func PhotoHashExists(hash string, ctx context.Context, client *Client) (bool, *Post, *UploadTask, error) {
	if postItem, err := client.Post.Query().Where(post.ImageHashEQ(hash)).First(ctx); !IsNotFound(err) {
		return true, postItem, nil, nil
	} else if uploadTaskItem, err := client.UploadTask.Query().Where(uploadtask.ImageHashEQ(hash)).First(ctx); !IsNotFound(err) {
		return true, nil, uploadTaskItem, nil
	} else if err != nil {
		return false, nil, nil, err
	}
	return false, nil, nil, nil
}
