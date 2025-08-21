package handlers

import (
	"channel-helper-go/utils"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func UnknownHandler(ctx *th.Context, message telego.Message) error {
	logger, _ := ctx.Value("logger").(utils.Logger)

	if gtfo(ctx, message) {
		return nil
	}

	logger.Info("UnknownHandler called")

	_, _ = ctx.Bot().SendMessage(ctx, tu.Message(
		message.Chat.ChatID(),
		"Unknown media type",
	))
	return nil
}
