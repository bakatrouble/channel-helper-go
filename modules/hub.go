package channels

import "channel-helper-go/ent"

type Hub struct {
	PostCreated       chan *ent.Post
	PostSent          chan *ent.Post
	PostDeleted       chan *ent.Post
	UploadTaskCreated chan *ent.UploadTask
	UploadTaskDone    chan *ent.UploadTask
}

func NewHub() Hub {
	return Hub{
		PostCreated:       make(chan *ent.Post, 100),
		PostSent:          make(chan *ent.Post, 100),
		PostDeleted:       make(chan *ent.Post, 100),
		UploadTaskCreated: make(chan *ent.UploadTask, 100),
		UploadTaskDone:    make(chan *ent.UploadTask, 100),
	}
}
