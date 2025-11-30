package handlers

import (
	"channel-helper-go/utils/cfg"
	"slices"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func reactToMessage(ctx *th.Context, message *telego.Message) {
	_ = ctx.Bot().SetMessageReaction(ctx, &telego.SetMessageReactionParams{
		ChatID:    message.Chat.ChatID(),
		MessageID: message.MessageID,
		Reaction: []telego.ReactionType{&telego.ReactionTypeEmoji{
			Type:  telego.ReactionEmoji,
			Emoji: "👍",
		}},
	})
}

func gtfo(ctx *th.Context, message telego.Message) bool {
	bot := ctx.Bot()
	config := ctx.Value("config").(*cfg.Config)

	if slices.Contains(config.AllowedSenderChats, message.Chat.ID) {
		return false
	}

	_, _ = bot.SendMessage(ctx, tu.Message(
		message.Chat.ChatID(),
		"GTFO",
	))
	return true
}
