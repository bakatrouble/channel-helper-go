package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func AnimationHandler(ctx *th.Context, message telego.Message) error {
	println("AnimationHandler called")
	db, _ := ctx.Value("db").(*ent.Client)

	createdPost, err := db.Post.Create().
		SetType(post.TypeAnimation).
		SetFileID(message.Animation.FileID).
		Save(ctx)
	if err != nil {
		println("Failed to create post:", err.Error())
		return err
	}
	_ = createPostMessageId(ctx, createdPost, &message)

	reactToMessage(ctx, &message)

	return nil
}
