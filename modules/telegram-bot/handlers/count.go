package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func CountHandler(ctx *th.Context, message telego.Message) error {
	db := ctx.Value("db").(*ent.Client)

	count, err := db.Post.Query().
		Where(post.IsSent(false)).
		Count(ctx)
	if err != nil {
		println("Failed to count posts:", err.Error())
		_, _ = ctx.Bot().SendMessage(ctx, tu.Message(
			message.Chat.ChatID(),
			"Error counting posts",
		))
		return err
	}

	// Send a response with the count
	responseText := fmt.Sprintf("Unsent count: %d", count)
	_, _ = ctx.Bot().SendMessage(ctx, tu.Message(
		message.Chat.ChatID(),
		responseText,
	))
	return nil
}
