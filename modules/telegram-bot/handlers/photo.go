package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"channel-helper-go/utils"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"time"
)

func PhotoHandler(ctx *th.Context, message telego.Message) error {
	db, _ := ctx.Value("db").(*ent.Client)
	hub, _ := ctx.Value("hub").(*utils.Hub)
	bot := ctx.Bot()
	logger, _ := ctx.Value("logger").(utils.Logger)

	logger.Info("PhotoHandler called")

	file, err := bot.GetFile(ctx, &telego.GetFileParams{FileID: message.Photo[len(message.Photo)-1].FileID})
	if err != nil {
		logger.With("err", err).Error("error getting photo")
		return nil
	}
	fileData, err := tu.DownloadFile(bot.FileDownloadURL(file.FilePath))
	if err != nil {
		logger.With("err", err).Error("error downloading photo")
		return nil
	}
	hash, err := utils.HashImage(fileData)
	if err != nil {
		logger.With("err", err).Error("error hashing photo")
		return nil
	}

	duplicate, dPost, dUploadTask, err := ent.PhotoHashExists(hash, ctx, db)
	if err != nil {
		logger.With("err", err).Error("error checking for duplicate photo hash")
		return nil
	}
	if duplicate {
		if dPost != nil {
			logger.With("hash", hash).With("post_id", dPost.ID).Info("duplicate photo hash found")
			_ = createPostMessageId(ctx, dPost, &message)

			newMsg, _ := bot.SendPhoto(ctx, tu.Photo(
				message.Chat.ChatID(),
				telego.InputFile{
					FileID: dPost.FileID,
				},
			).
				WithCaption(
					fmt.Sprintf("Duplicate from %s", dPost.CreatedAt.Format(time.RFC3339)),
				).
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
					ChatID:    message.Chat.ChatID(),
				}))
			if newMsg != nil {
				_ = createPostMessageId(ctx, dPost, newMsg)
			}
			return nil
		} else if dUploadTask != nil {
			logger.With("hash", hash).With("task_id", dUploadTask.ID).Info("duplicate upload task found")
			_, _ = bot.SendMessage(ctx, tu.Message(
				message.Chat.ChatID(),
				fmt.Sprintf("Duplicate upload task from %s", dUploadTask.CreatedAt.Format(time.RFC3339)),
			).
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
					ChatID:    message.Chat.ChatID(),
				}))
			return nil
		}
	}

	createdPost, err := db.Post.Create().
		SetType(post.TypePhoto).
		SetFileID(message.Photo[len(message.Photo)-1].FileID).
		SetImageHash(hash).
		Save(ctx)
	if err != nil {
		logger.With("err", err).Error("failed to create post")
		return nil
	}
	_ = createPostMessageId(ctx, createdPost, &message)

	reactToMessage(ctx, &message)

	hub.PostCreated <- createdPost
	logger.With("id", createdPost.ID).Info("created photo post id")

	return nil
}
