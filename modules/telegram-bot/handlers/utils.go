package handlers

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func reactToMessage(ctx *th.Context, message telego.Message) {
	_ = ctx.Bot().SetMessageReaction(ctx, &telego.SetMessageReactionParams{
		ChatID:    message.Chat.ChatID(),
		MessageID: message.MessageID,
		Reaction: []telego.ReactionType{&telego.ReactionTypeEmoji{
			Type:  telego.ReactionEmoji,
			Emoji: "👍",
		}},
	})
}
