package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"channel-helper-go/ent/postmessageid"
	"channel-helper-go/utils"
	"errors"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func DeleteCommandHandler(ctx *th.Context, message telego.Message) error {
	logger, _ := ctx.Value("logger").(utils.Logger)
	bot := ctx.Bot()

	logger.Info("DeleteCommandHandler called")

	replyParameters := &telego.ReplyParameters{
		MessageID: message.MessageID,
		ChatID:    message.Chat.ChatID(),
	}

	if message.ReplyToMessage == nil {
		_, _ = bot.SendMessage(ctx, tu.Message(
			message.Chat.ChatID(),
			"Please reply to a message with media to delete it",
		))
		return nil
	}

	logicErr, err := deleteByMessage(ctx, message.ReplyToMessage)
	if logicErr != nil {
		_, _ = bot.SendMessage(ctx, tu.Message(
			message.Chat.ChatID(),
			logicErr.Error(),
		).WithReplyParameters(replyParameters))
		return nil
	} else if err != nil {
		_, _ = bot.SendMessage(ctx, tu.Message(
			message.Chat.ChatID(),
			"An error occurred while trying to delete the post",
		).WithReplyParameters(replyParameters))
		return err
	}

	reactToMessage(ctx, &message)

	return nil
}

func DeleteCallbackHandler(ctx *th.Context, callbackQuery telego.CallbackQuery) error {
	logger, _ := ctx.Value("logger").(utils.Logger)
	bot := ctx.Bot()

	logger.Info("DeleteCallbackHandler called")

	if !callbackQuery.Message.IsAccessible() {
		return nil
	}
	logicErr, err := deleteByMessage(ctx, callbackQuery.Message.Message())
	if logicErr != nil {
		err = bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(callbackQuery.ID).
			WithText(logicErr.Error()),
		)
		return nil
	} else if err != nil {
		err = bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(callbackQuery.ID).
			WithText("An error has occurred"))
		return err
	}
	_, _ = bot.EditMessageReplyMarkup(ctx, &telego.EditMessageReplyMarkupParams{
		ChatID:      callbackQuery.Message.GetChat().ChatID(),
		MessageID:   callbackQuery.Message.GetMessageID(),
		ReplyMarkup: tu.InlineKeyboard(),
	})

	return nil
}

func deleteByMessage(ctx *th.Context, message *telego.Message) (error, error) {
	db := ctx.Value("db").(*ent.Client)
	hub, _ := ctx.Value("hub").(*utils.Hub)
	logger, _ := ctx.Value("logger").(utils.Logger)

	logger.With("chat_id", message.Chat.ID, "message_id", message.MessageID).Info("deleting post")

	postObj, err := db.Post.Query().
		Where(post.HasMessageIdsWith(
			postmessageid.ChatID(message.Chat.ID),
			postmessageid.MessageID(message.MessageID),
		)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errors.New("post not found"), nil
		}
		logger.With("err", err).Error("failed to query post")
		return nil, err
	}

	err = db.Post.DeleteOne(postObj).Exec(ctx)
	if err != nil {
		logger.With("err", err).Error("failed to delete post")
		return nil, err
	}

	hub.PostDeleted <- postObj
	logger.With("id", postObj.ID).Info("deleted post")

	return nil, nil
}
