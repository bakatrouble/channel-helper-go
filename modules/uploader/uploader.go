package uploader

import (
	"channel-helper-go/database"
	"channel-helper-go/database/database_utils"
	telegram_bot "channel-helper-go/modules/telegram-bot"
	"channel-helper-go/utils"
	"context"
	"database/sql"
	"errors"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"sync"
	"time"
)

func processTask(task *database.UploadTask, bot *telego.Bot, ctx context.Context) error {
	config := ctx.Value("config").(*utils.Config)
	db := ctx.Value("db").(*database.DBStruct)
	hub := ctx.Value("hub").(*utils.Hub)
	logger := ctx.Value("logger").(utils.Logger)
	var err error

	post := &database.Post{
		Type:      task.Type,
		ImageHash: task.ImageHash,
	}
	replyMarkup := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("Delete").
				WithCallbackData("/delete"),
		),
	)
	var msg *telego.Message
	switch task.Type {
	case database.MediaTypePhoto:
		msg, err = bot.SendPhoto(ctx, &telego.SendPhotoParams{
			ChatID:      telego.ChatID{ID: config.UploadChatId},
			Photo:       tu.FileFromBytes(*task.Data, "image.jpg"),
			ReplyMarkup: replyMarkup,
		})
		if err != nil {
			logger.With("err", err).Error("error uploading photo")
			return err
		}

		post.FileID = msg.Photo[len(msg.Photo)-1].FileID
	case database.MediaTypeAnimation:
		msg, err = bot.SendAnimation(ctx, &telego.SendAnimationParams{
			ChatID:      telego.ChatID{ID: config.UploadChatId},
			Animation:   tu.FileFromBytes(*task.Data, "image.gif"),
			ReplyMarkup: replyMarkup,
		})
		if err != nil {
			logger.With("err", err).Error("error uploading animation")
			return err
		}

		post.FileID = msg.Animation.FileID
	default:
		logger.With("type", task.Type).Error("unsupported upload task type")
		return errors.New("unsupported upload task type")
	}

	post.MessageIDs = []database.MessageID{
		{
			ChatID:    msg.Chat.ID,
			MessageID: msg.MessageID,
		},
	}
	err = db.Post.Create(ctx, post)
	if err != nil {
		logger.With("err", err).Error("error creating post")
		return err
	}

	task.SentAt = database_utils.Now()
	task.IsProcessed = true
	task.Data = nil
	task.ImageHash = nil
	err = db.UploadTask.Update(ctx, task)
	if err != nil {
		logger.With("err", err).Error("error updating upload task")
		return err
	}

	hub.UploadTaskDone <- task
	hub.PostCreated <- post
	logger.With("post_id", post.ID).
		With("task_id", task.ID).
		Info("created post from upload task")

	return nil
}

func StartUploader(ctx context.Context) {
	wg := ctx.Value("wg").(*sync.WaitGroup)
	hub := ctx.Value("hub").(*utils.Hub)
	sqldb := ctx.Value("sqldb").(*sql.DB)
	config := ctx.Value("config").(*utils.Config)

	defer wg.Done()

	logger := utils.NewLogger(config.DbName, "uploader")
	ctx = context.WithValue(ctx, "logger", logger)
	logger.Info("starting uploader")

	db, err := database.NewDBStruct(sqldb, !config.Production, logger)
	if err != nil {
		logger.With("err", err).Error("failed to connect to database")
		panic(err)
	}
	defer func(db *database.DBStruct) {
		_ = db.Close()
	}(db)
	ctx = context.WithValue(ctx, "db", db)

	bot, err := telegram_bot.CreateBot(ctx, logger)
	if err != nil {
		logger.With("err", err).Error("failed to create bot")
		return
	}

	putUnprocessedTasks := func() {
		tasks, err := db.UploadTask.GetUnsent(ctx)
		if err != nil {
			logger.With("err", err).Error("initial upload tasks fetch error")
			return
		}
		for _, task := range tasks {
			hub.UploadTaskCreated <- task
		}
		logger.With("count", len(tasks)).Info("upload tasks remaining")
	}

	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case task := <-hub.UploadTaskCreated:
			_ = processTask(task, bot, ctx)
			logger.With("count", len(hub.UploadTaskCreated)).Info("upload tasks remaining")
		case <-ticker.C:
			putUnprocessedTasks()
			ticker.Reset(time.Minute)
		case <-ctx.Done():
			return
		}
	}
}
