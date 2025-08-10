package handlers

import (
	"channel-helper-go/database"
	"channel-helper-go/utils"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func VideoHandler(ctx *th.Context, message telego.Message) error {
	db, _ := ctx.Value("db").(*database.DBStruct)
	hub, _ := ctx.Value("hub").(*utils.Hub)
	logger, _ := ctx.Value("logger").(utils.Logger)

	logger.Info("VideoHandler called")

	post := &database.Post{
		Type:   database.MediaTypeVideo,
		FileID: message.Video.FileID,
		MessageIDs: []*database.MessageID{
			{
				ChatID:    message.Chat.ID,
				MessageID: message.MessageID,
			},
		},
	}
	err := db.Post.Create(ctx, post)
	if err != nil {
		logger.With("err", err).Error("failed to create post")
		return err
	}

	reactToMessage(ctx, &message)

	hub.PostCreated <- post

	return nil
}
