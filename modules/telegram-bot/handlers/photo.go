package handlers

import (
	"channel-helper-go/database"
	"channel-helper-go/utils"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"time"
)

func PhotoHandler(ctx *th.Context, message telego.Message) error {
	db, _ := ctx.Value("db").(*database.DBStruct)
	hub, _ := ctx.Value("hub").(*utils.Hub)
	bot := ctx.Bot()
	logger, _ := ctx.Value("logger").(utils.Logger)

	logger.Info("PhotoHandler called")

	file, err := bot.GetFile(ctx, &telego.GetFileParams{FileID: message.Photo[len(message.Photo)-1].FileID})
	if err != nil {
		logger.With("err", err).Error("error getting image")
		return nil
	}
	fileData, err := tu.DownloadFile(bot.FileDownloadURL(file.FilePath))
	if err != nil {
		logger.With("err", err).Error("error downloading image")
		return nil
	}
	hash, err := utils.HashImage(fileData)
	if err != nil {
		logger.With("err", err).Error("error hashing image")
		return nil
	}
	logger.With("hash", hash).Info("image hash calculated")

	duplicate, dPost, dUploadTask, err := db.ImageHash.Exists(ctx, hash)
	if err != nil {
		logger.With("err", err).Error("error checking for duplicate image hash")
		return nil
	}
	if duplicate {
		if dPost != nil {
			logger.With("hash", hash).With("post_id", dPost.ID).Info("duplicate photo hash found")
			dPost.MessageIDs = append(dPost.MessageIDs, database.MessageID{
				ChatID:    message.Chat.ID,
				MessageID: message.MessageID,
			})
			err = db.Post.Update(ctx, dPost)
			if err != nil {
				logger.With("err", err).Error("error updating post with new message ID")
			}

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
				dPost.MessageIDs = append(dPost.MessageIDs, database.MessageID{
					ChatID:    message.Chat.ID,
					MessageID: message.MessageID,
				})
				err = db.Post.Update(ctx, dPost)
				if err != nil {
					logger.With("err", err).Error("error updating post with new message ID")
				}
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

	post := &database.Post{
		Type:   database.MediaTypePhoto,
		FileID: message.Photo[len(message.Photo)-1].FileID,
		MessageIDs: []database.MessageID{
			{
				ChatID:    message.Chat.ID,
				MessageID: message.MessageID,
			},
		},
		ImageHash: &database.ImageHash{
			Hash: hash,
		},
	}
	err = db.Post.Create(ctx, post)
	reactToMessage(ctx, &message)

	hub.PostCreated <- post
	logger.With("id", post.ID).Info("created photo post id")

	return nil
}
