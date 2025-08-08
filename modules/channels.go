package channels

import "channel-helper-go/ent"

type AppChannels struct {
	UploadTaskCreated chan *ent.UploadTask
	PostCreated       chan *ent.Post
	UploadTaskDone    chan *ent.UploadTask
	PostSent          chan *ent.Post
	PostDeleted       chan *ent.Post
}

func NewAppChannels() AppChannels {
	return AppChannels{
		UploadTaskCreated: make(chan *ent.UploadTask, 100),
		PostCreated:       make(chan *ent.Post, 100),
		UploadTaskDone:    make(chan *ent.UploadTask, 100),
		PostSent:          make(chan *ent.Post, 100),
	}
}
