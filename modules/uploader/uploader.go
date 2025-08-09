package uploader

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"channel-helper-go/ent/uploadtask"
	telegram_bot "channel-helper-go/modules/telegram-bot"
	"channel-helper-go/utils"
	"context"
	"errors"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"sync"
	"time"
)

func processTask(task *ent.UploadTask, bot *telego.Bot, ctx context.Context) error {
	config := ctx.Value("config").(*utils.Config)
	db := ctx.Value("db").(*ent.Client)
	hub := ctx.Value("hub").(*utils.Hub)
	logger := ctx.Value("logger").(utils.Logger)

	tx, err := db.Tx(ctx)
	if err != nil {
		logger.With("err", err).Error("error starting transaction")
		return err
	}

	postBuilder := tx.Post.Create()
	if task.Edges.ImageHash != nil {
		postBuilder.SetImageHash(task.Edges.ImageHash)
	}
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
			logger.With("err", err).Error("error uploading photo")
			return err
		}

		createdPost, err = postBuilder.
			SetType(post.TypePhoto).
			SetFileID(msg.Photo[len(msg.Photo)-1].FileID).
			Save(ctx)
		if err != nil {
			logger.With("err", err).Error("error creating post")
			return err
		}
	case uploadtask.TypeAnimation:
		msg, err = bot.SendAnimation(ctx, &telego.SendAnimationParams{
			ChatID:      telego.ChatID{ID: config.UploadChatId},
			Animation:   tu.FileFromBytes(task.Data, "image.gif"),
			ReplyMarkup: replyMarkup,
		})
		if err != nil {
			logger.With("err", err).Error("error uploading animation")
			return err
		}

		createdPost, err = postBuilder.
			SetType(post.TypeAnimation).
			SetFileID(msg.Animation.FileID).
			Save(ctx)
		if err != nil {
			logger.With("err", err).Error("error creating post")
			return err
		}
	default:
		logger.With("type", task.Type).Error("unsupported upload task type")
		return errors.New("unsupported upload task type")
	}

	err = tx.UploadTask.UpdateOne(task).
		SetSentAt(time.Now()).
		SetIsProcessed(true).
		SetData([]byte{}).
		Exec(ctx)
	if err != nil {
		logger.With("err", err).Error("error updating upload task")
		return err
	}

	err = tx.PostMessageId.Create().
		SetMessageID(msg.MessageID).
		SetChatID(config.UploadChatId).
		SetPost(createdPost).
		Exec(ctx)
	if err != nil {
		logger.With("err", err).Error("error creating post message ID")
		return err
	}

	err = tx.Commit()
	if err != nil {
		logger.With("err", err).Error("error committing transaction")
		return err
	}

	hub.UploadTaskDone <- task
	hub.PostCreated <- createdPost
	logger.With("post_id", createdPost.ID).
		With("task_id", task.ID).
		Info("created post from upload task")

	return nil
}

func StartUploader(ctx context.Context) {
	wg := ctx.Value("wg").(*sync.WaitGroup)
	db := ctx.Value("db").(*ent.Client)
	hub := ctx.Value("hub").(*utils.Hub)
	config := ctx.Value("config").(*utils.Config)

	defer wg.Done()

	logger := utils.NewLogger(config.DbName, "uploader")
	ctx = context.WithValue(ctx, "logger", logger)
	logger.Info("starting uploader")

	bot, err := telegram_bot.CreateBot(ctx, logger)
	if err != nil {
		logger.With("err", err).Error("failed to create bot")
		return
	}

	tasks, err := db.UploadTask.Query().
		Where(uploadtask.IsProcessed(false)).
		All(ctx)
	if err != nil {
		logger.With("err", err).Error("initial upload tasks fetch error")
		return
	}
	for _, task := range tasks {
		hub.UploadTaskCreated <- task
	}
	logger.With("count", len(tasks)).Info("upload tasks remaining")

	for {
		select {
		case task := <-hub.UploadTaskCreated:
			_ = processTask(task, bot, ctx)
			logger.With("count", len(hub.UploadTaskCreated)).Info("upload tasks remaining")
		case <-ctx.Done():
			return
		}
	}
}
