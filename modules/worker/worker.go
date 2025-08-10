package worker

import (
	"channel-helper-go/database"
	"channel-helper-go/database/database_utils"
	telegram_bot "channel-helper-go/modules/telegram-bot"
	"channel-helper-go/utils"
	"context"
	"database/sql"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"sync"
	"time"
)

func sendPost(post *database.Post, bot *telego.Bot, ctx context.Context) error {
	config := ctx.Value("config").(*utils.Config)
	hub := ctx.Value("hub").(*utils.Hub)
	logger := ctx.Value("logger").(utils.Logger)
	db := ctx.Value("db").(*database.DBStruct)
	var err error

	logger.With("id", post.ID).Info("sending post")

	switch post.Type {
	case database.MediaTypePhoto:
		_, err = bot.SendPhoto(ctx, tu.Photo(
			telego.ChatID{ID: config.TargetChatId},
			telego.InputFile{
				FileID: post.FileID,
			},
		))
	case database.MediaTypeVideo:
		_, err = bot.SendVideo(ctx, tu.Video(
			telego.ChatID{ID: config.TargetChatId},
			telego.InputFile{
				FileID: post.FileID,
			},
		))
	case database.MediaTypeAnimation:
		_, err = bot.SendAnimation(ctx, tu.Animation(
			telego.ChatID{ID: config.TargetChatId},
			telego.InputFile{
				FileID: post.FileID,
			},
		))
	}

	if err != nil {
		logger.With("err", err).Error("error sending post")
		return err
	}

	post.IsSent = true
	post.SentAt = database_utils.Now()
	err = db.Post.Update(ctx, post)
	if err != nil {
		logger.With("err", err).Error("error updating post as sent")
		return err
	}

	hub.PostSent <- post

	logger.With("id", post.ID).Info("sent post")

	return nil
}

func unsentPostsCount(ctx context.Context) int {
	db := ctx.Value("db").(*database.DBStruct)
	cnt, _ := db.Post.UnsentCount(ctx)
	return cnt
}

func fetchAndSendPost(bot *telego.Bot, ctx context.Context) {
	config := ctx.Value("config").(*utils.Config)
	db := ctx.Value("db").(*database.DBStruct)
	logger := ctx.Value("logger").(utils.Logger)

	postObj, err := db.Post.GetRandomUnsent(ctx)
	if err != nil {
		logger.With("err", err).Error("error fetching posts")
		return
	}
	if postObj == nil {
		logger.With("wait", config.Interval).Info("no unsent posts found")
		return
	}
	_ = sendPost(postObj, bot, ctx)
	logger.With("count", unsentPostsCount(ctx)).
		With("wait", config.Interval).
		Info("unsent posts remaining")
}

func StartWorker(ctx context.Context) {
	config := ctx.Value("config").(*utils.Config)
	sqldb := ctx.Value("sqldb").(*sql.DB)
	wg := ctx.Value("wg").(*sync.WaitGroup)

	defer wg.Done()

	logger := utils.NewLogger(config.DbName, "worker")
	ctx = context.WithValue(ctx, "logger", logger)

	db, err := database.NewDBStruct(sqldb, !config.Production, logger)
	if err != nil {
		logger.With("err", err).Error("failed to connect to database")
		panic(err)
	}
	defer func(db *database.DBStruct) {
		_ = db.Close()
	}(db)
	ctx = context.WithValue(ctx, "db", db)

	logger.Info("starting worker")
	logger.With("count", unsentPostsCount(ctx)).With("wait", config.Interval).Info("unsent posts remaining")

	bot, err := telegram_bot.CreateBot(ctx, logger)
	if err != nil {
		logger.With("err", err).Error("failed to create bot")
		return
	}

	fetchAndSendPost(bot, ctx)
	ticker := time.NewTimer(config.Interval)
	for {
		select {
		case <-ticker.C:
			fetchAndSendPost(bot, ctx)
			ticker.Reset(config.Interval)
		case <-ctx.Done():
			logger.Info("stopping worker")
			return
		}
	}
}
