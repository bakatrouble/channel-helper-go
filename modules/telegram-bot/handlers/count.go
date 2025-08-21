package handlers

import (
	"channel-helper-go/database"
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func CountHandler(ctx *th.Context, message telego.Message) error {
	db := ctx.Value("db").(*database.DBStruct)

	if gtfo(ctx, message) {
		return nil
	}

	count, err := db.Post.UnsentCount(ctx)
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
