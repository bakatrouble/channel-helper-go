package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"channel-helper-go/utils"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func AnimationHandler(ctx *th.Context, message telego.Message) error {
	db, _ := ctx.Value("db").(*ent.Client)
	hub, _ := ctx.Value("hub").(*utils.Hub)
	logger, _ := ctx.Value("logger").(utils.Logger)

	logger.Info("AnimationHandler called")

	createdPost, err := db.Post.Create().
		SetType(post.TypeAnimation).
		SetFileID(message.Animation.FileID).
		Save(ctx)
	if err != nil {
		logger.With("err", err).Error("failed to create post")
		return err
	}
	_ = createPostMessageId(ctx, createdPost, &message)

	reactToMessage(ctx, &message)

	hub.PostCreated <- createdPost
	logger.With("id", createdPost.ID).Info("created animation post")

	return nil
}
