package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"channel-helper-go/utils"
	"errors"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"time"
)

func PhotoHandler(ctx *th.Context, message telego.Message) error {
	println("PhotoHandler called")
	db, _ := ctx.Value("db").(*ent.Client)
	bot := ctx.Bot()

	file, err := bot.GetFile(ctx, &telego.GetFileParams{FileID: message.Photo[len(message.Photo)-1].FileID})
	if err != nil {
		println("Error getting photo:", err.Error())
		return err
	}
	fileData, err := tu.DownloadFile(bot.FileDownloadURL(file.FilePath))
	if err != nil {
		println("Error downloading photo:", err.Error())
		return err
	}
	hash, err := utils.HashImage(fileData)
	if err != nil {
		println("Error hashing photo:", err.Error())
		return err
	}

	duplicate, dPost, dUploadTask, err := ent.PhotoHashExists(hash, ctx, db)
	if err != nil {
		println("Error checking for duplicate photo hash:", err.Error())
		return err
	}
	if duplicate {
		if dPost != nil {
			_, _ = bot.SendPhoto(ctx, tu.Photo(
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
			return errors.New("duplicate image hash")
		} else if dUploadTask != nil {
			_, _ = bot.SendMessage(ctx, tu.Message(
				message.Chat.ChatID(),
				fmt.Sprintf("Duplicate upload task from %s", dUploadTask.CreatedAt.Format(time.RFC3339)),
			).
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
					ChatID:    message.Chat.ChatID(),
				}))
			return errors.New("duplicate upload task")
		}
	}

	err = db.Post.Create().
		SetType(post.TypePhoto).
		SetFileID(message.Photo[len(message.Photo)-1].FileID).
		SetImageHash(hash).
		Exec(ctx)
	if err != nil {
		println("Failed to create post:", err.Error())
		return err
	}

	reactToMessage(ctx, message)

	return nil
}
