package handlers

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func UnknownHandler(ctx *th.Context, message telego.Message) error {
	_, _ = ctx.Bot().SendMessage(ctx, tu.Message(
		message.Chat.ChatID(),
		"Unknown media type",
	))
	return nil
}
