package utils

import (
	"channel-helper-go/database"
)

type Hub struct {
	PostCreated       chan *database.Post
	PostSent          chan *database.Post
	PostDeleted       chan *database.Post
	UploadTaskCreated chan *database.UploadTask
	UploadTaskDone    chan *database.UploadTask
}

func NewHub() Hub {
	return Hub{
		PostCreated:       make(chan *database.Post, 1024),
		PostSent:          make(chan *database.Post, 1024),
		PostDeleted:       make(chan *database.Post, 1024),
		UploadTaskCreated: make(chan *database.UploadTask, 1024),
		UploadTaskDone:    make(chan *database.UploadTask, 1024),
	}
}
