package worker

import (
	"channel-helper-go/database"
	"channel-helper-go/database/database_utils"
	telegram_bot "channel-helper-go/modules/telegram-bot"
	"channel-helper-go/utils"
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func sendPost(post *database.Post, bot *telego.Bot, ctx context.Context) error {
	config := ctx.Value("config").(*utils.Config)
	hub := ctx.Value("hub").(*utils.Hub)
	logger := ctx.Value("logger").(utils.Logger)
	db := ctx.Value("db").(*database.DBStruct)
	var err error

	logger.With("id", post.ID).Info("sending post")

	processedPosts := make([]*database.Post, 1)
	processedPosts[0] = post

	chatId := telego.ChatID{ID: config.TargetChatId}
	inputFile := telego.InputFile{FileID: post.FileID}
	switch post.Type {
	case database.MediaTypePhoto:
		unsentCount := unsentPostsCount(ctx)
		if unsentCount > config.GroupThreshold {
			logger.With("count", unsentCount).Info("grouping photos for media group")
			extraPosts, _ := db.Post.GetAdditionalUnsentByType(ctx, database.MediaTypePhoto, post.ID)
			media := make([]telego.InputMedia, 1)
			media[0] = tu.MediaPhoto(inputFile)
			for _, extraPost := range extraPosts {
				media = append(media, tu.MediaPhoto(telego.InputFile{FileID: extraPost.FileID}))
			}
			if len(media) > 1 {
				_, err = bot.SendMediaGroup(ctx, tu.MediaGroup(
					chatId,
					media...,
				))
				for _, extraPost := range extraPosts {
					processedPosts = append(processedPosts, extraPost)
				}
				break
			}
		}
		_, err = bot.SendPhoto(ctx, tu.Photo(chatId, inputFile))
	case database.MediaTypeVideo:
		_, err = bot.SendVideo(ctx, tu.Video(chatId, inputFile))
	case database.MediaTypeAnimation:
		_, err = bot.SendAnimation(ctx, tu.Animation(chatId, inputFile))
	}

	if err != nil {
		logger.With("err", err).Error("error sending post")
		return err
	}

	for _, p := range processedPosts {
		p.IsSent = true
		p.SentAt = database_utils.Now()
		err = db.Post.Update(ctx, p)
		if err != nil {
			logger.With("err", err).Error("error updating post as sent")
			return err
		}
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

	post, err := db.Post.GetRandomUnsent(ctx)
	if err != nil {
		logger.With("err", err).Error("error fetching posts")
		return
	}
	if post == nil {
		logger.With("wait", config.Interval).Info("no unsent posts found")
		return
	}
	_ = sendPost(post, bot, ctx)
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

	//fetchAndSendPost(bot, ctx)
	ticker := time.NewTimer(config.Interval)
	for {
		select {
		case <-ticker.C:
			fetchAndSendPost(bot, ctx)
		case <-ctx.Done():
			logger.Info("stopping worker")
			return
		}
	}
}
