package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/utils"
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
	logger, _ := ctx.Value("logger").(utils.Logger)

	_, err := db.PostMessageId.Create().
		SetPost(post).
		SetMessageID(message.MessageID).
		SetChatID(message.Chat.ID).
		Save(ctx)
	if err != nil {
		logger.With("err", err).Error("failed to create PostMessageId")
		return err
	}
	return nil
}
