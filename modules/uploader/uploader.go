package uploader

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"channel-helper-go/ent/uploadtask"
	channels "channel-helper-go/modules"
	"channel-helper-go/utils"
	"context"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"sync"
	"time"
)

func processTask(task *ent.UploadTask, ctx context.Context) error {
	config := ctx.Value("config").(*utils.Config)
	db := ctx.Value("db").(*ent.Client)
	bot := ctx.Value("bot").(*telego.Bot)
	chans := c.MustGet("chans").(*channels.AppChannels)

	tx, err := db.Tx(ctx)
	if err != nil {
		println("Error starting transaction:", err.Error())
		return err
	}

	postBuilder := tx.Post.Create()
	replyMarkup := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Delete").
				WithCallbackData("/delete"),
		),
	)
	var msg *telego.Message
	var createdPost *ent.Post
	switch task.Type {
	case uploadtask.TypePhoto:
		msg, err = bot.SendPhoto(ctx, &telego.SendPhotoParams{
			ChatID:      telego.ChatID{ID: config.UploadChatId},
			Photo:       tu.FileFromBytes(task.Data, "image.jpg"),
			ReplyMarkup: replyMarkup,
		})
		if err != nil {
			println("Error uploading photo:", err.Error())
			return err
		}

		createdPost, err = postBuilder.
			SetType(post.TypePhoto).
			SetFileID(msg.Animation.FileID).
			SetImageHash(task.ImageHash).
			Save(ctx)
		if err != nil {
			println("Error creating post:", err.Error())
			return err
		}
	case uploadtask.TypeAnimation:
		msg, err = bot.SendAnimation(ctx, &telego.SendAnimationParams{
			ChatID:      telego.ChatID{ID: config.UploadChatId},
			Animation:   tu.FileFromBytes(task.Data, "image.gif"),
			ReplyMarkup: replyMarkup,
		})
		if err != nil {
			println("Error uploading photo:", err.Error())
			return err
		}

		createdPost, err = postBuilder.
			SetType(post.TypeAnimation).
			SetFileID(msg.Animation.FileID).
			Save(ctx)
		if err != nil {
			println("Error creating post:", err.Error())
			return err
		}
	default:
		println("Unsupported upload task type:", task.Type)
		return err
	}

	err = tx.UploadTask.UpdateOne(task).
		SetSentAt(time.Now()).
		SetIsProcessed(true).
		SetData([]byte{}).
		Exec(ctx)
	if err != nil {
		println("Error updating upload task:", err.Error())
		return err
	}

	err = tx.PostMessageId.Create().
		SetMessageID(msg.MessageID).
		SetChatID(config.UploadChatId).
		SetPost(createdPost).
		Exec(ctx)
	if err != nil {
		println("Error creating post message ID:", err.Error())
		return err
	}

	err = tx.Commit()
	if err != nil {
		println("Error committing transaction:", err.Error())
		return err
	}

	chans.UploadTaskDone <- task
	chans.PostCreated <- createdPost

	return nil
}

func StartUploader(ctx context.Context) {
	wg := ctx.Value("wg").(*sync.WaitGroup)
	db := ctx.Value("db").(*ent.Client)
	uploadTaskChannel := ctx.Value("uploadTaskChannel").(chan *ent.UploadTask)

	defer wg.Done()

	tasks, err := db.UploadTask.Query().
		Where(uploadtask.IsProcessed(false)).
		All(ctx)
	if err != nil {
		println("Initial upload tasks fetch error:", err.Error())
		return
	}
	for _, task := range tasks {
		uploadTaskChannel <- task
	}

	for {
		select {
		case <-ctx.Done():
			return
		case task := <-uploadTaskChannel:
			if task != nil {
				_ = processTask(task, ctx)
			}
		default:
			// Continue processing tasks
		}
	}
}
