package handlers

import (
	"channel-helper-go/ent"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
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

func createPostMessageId(ctx *th.Context, post *ent.Post, message *telego.Message) error {
	db, _ := ctx.Value("db").(*ent.Client)
	_, err := db.PostMessageId.Create().
		SetPost(post).
		SetMessageID(message.MessageID).
		SetChatID(message.Chat.ID).
		Save(ctx)
	if err != nil {
		println("Failed to create PostMessageId:", err.Error())
		return err
	}
	return nil
}
