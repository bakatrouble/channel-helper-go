package handlers

import (
	"channel-helper-go/database"
	"channel-helper-go/utils"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func AnimationHandler(ctx *th.Context, message telego.Message) error {
	db, _ := ctx.Value("db").(*database.DBStruct)
	hub, _ := ctx.Value("hub").(*utils.Hub)
	logger, _ := ctx.Value("logger").(utils.Logger)

	if gtfo(ctx, message) {
		return nil
	}

	logger.Info("AnimationHandler called")

	post := &database.Post{
		Type:   database.MediaTypeAnimation,
		FileID: message.Animation.FileID,
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
	logger.With("id", post.ID).Info("created animation post")

	return nil
}
