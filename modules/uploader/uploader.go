package uploader

import (
	"channel-helper-go/database"
	"channel-helper-go/database/database_utils"
	telegram_bot "channel-helper-go/modules/telegram-bot"
	"channel-helper-go/utils"
	"context"
	"database/sql"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func processTask(taskId string, bot *telego.Bot, ctx context.Context) error {
	config := ctx.Value("config").(*utils.Config)
	db := ctx.Value("db").(*database.DBStruct)
	//hub := ctx.Value("hub").(*utils.Hub)
	logger := ctx.Value("logger").(utils.Logger)
	var err error

	task, err := db.UploadTask.GetByID(ctx, taskId)
	if err != nil {
		logger.With("err", err).Error("error fetching upload task by ID")
		return err
	}
	if task == nil {
		return errors.New("task not found during refetch")
	}
	if task.IsProcessed {
		logger.With("task_id", task.ID).Info("task already processed, skipping")
		return nil
	}

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
		msg, err = bot.SendDocument(ctx, &telego.SendDocumentParams{
			ChatID:      telego.ChatID{ID: config.UploadChatId},
			Document:    tu.FileFromBytes(*task.Data, "image.gif"),
			ReplyMarkup: replyMarkup,
		})
		if err != nil {
			logger.With("err", err).Error("error uploading animation")
			return err
		}

		post.FileID = msg.Document.FileID
	default:
		logger.With("type", task.Type).Error("unsupported upload task type")
		return errors.New("unsupported upload task type")
	}

	post.MessageIDs = []*database.MessageID{
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

	//hub.UploadTaskDone <- task
	//hub.PostCreated <- post
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

	queue := make([]string, 0)

	putUnprocessedTasks := func() {
		logger.Info("getting unprocessed upload tasks")
		tasks, err := db.UploadTask.GetUnsent(ctx)
		if err != nil {
			logger.With("err", err).Error("initial upload tasks fetch error")
			return
		}
		for _, task := range tasks {
			if !slices.Contains(queue, task.ID) {
				queue = append(queue, task.ID)
			}
		}
		logger.With("count", len(tasks)).Info("upload tasks remaining")
	}

	channelWatcher := func() {
		for {
			task := <-hub.UploadTaskCreated
			queue = append(queue, task.ID)
			logger.With("count", len(queue)).Info("upload tasks created")
		}
	}

	retryAfterRegex, _ := regexp.Compile(`retry after (\d+)`)
	go putUnprocessedTasks()
	go channelWatcher()
	uploadTicker := time.NewTicker(time.Second)
	backlogTicker := time.NewTicker(time.Minute)
	for {
		select {
		case <-uploadTicker.C:
			if len(queue) > 0 {
				taskId := queue[0]
				queue = queue[1:]
				err = processTask(taskId, bot, ctx)
				logger.With("count", len(hub.UploadTaskCreated)).Info("upload tasks remaining")
				if err != nil {
					if matched := retryAfterRegex.FindStringSubmatch(err.Error()); matched != nil {
						seconds, _ := strconv.Atoi(matched[1])
						uploadTicker.Reset(time.Second * time.Duration(seconds))
						continue
					}
				}
			}
			uploadTicker.Reset(time.Second)
		case <-backlogTicker.C:
			putUnprocessedTasks()
			logger.Info("filled upload task backlog")
		case <-ctx.Done():
			return
		}
	}
}
