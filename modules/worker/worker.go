package worker

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	telegram_bot "channel-helper-go/modules/telegram-bot"
	"channel-helper-go/utils"
	"context"
	"entgo.io/ent/dialect/sql"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"sync"
	"time"
)

func sendPost(postObj *ent.Post, bot *telego.Bot, ctx context.Context) error {
	config := ctx.Value("config").(*utils.Config)
	hub := ctx.Value("hub").(*utils.Hub)
	logger := ctx.Value("logger").(utils.Logger)
	var err error

	logger.With("id", postObj.ID).Info("sending post")

	switch postObj.Type {
	case post.TypePhoto:
		_, err = bot.SendPhoto(ctx, tu.Photo(
			telego.ChatID{ID: config.TargetChatId},
			telego.InputFile{
				FileID: postObj.FileID,
			},
		))
	case post.TypeVideo:
		_, err = bot.SendVideo(ctx, tu.Video(
			telego.ChatID{ID: config.TargetChatId},
			telego.InputFile{
				FileID: postObj.FileID,
			},
		))
	case post.TypeAnimation:
		_, err = bot.SendAnimation(ctx, tu.Animation(
			telego.ChatID{ID: config.TargetChatId},
			telego.InputFile{
				FileID: postObj.FileID,
			},
		))
	}

	if err != nil {
		logger.With("err", err).Error("error sending post")
		return err
	}

	err = postObj.Update().
		SetIsSent(true).
		SetSentAt(time.Now()).
		Exec(ctx)
	if err != nil {
		logger.With("err", err).Error("error updating post as sent")
		return err
	}

	hub.PostSent <- postObj

	logger.With("id", postObj.ID).Info("sent post")

	return nil
}

func unsentPostsCount(ctx context.Context) int {
	db := ctx.Value("db").(*ent.Client)
	cnt, _ := db.Post.Query().
		Where(post.IsSent(false)).
		Order(sql.OrderByRand()).
		Count(ctx)
	return cnt
}

func StartWorker(ctx context.Context) {
	db := ctx.Value("db").(*ent.Client)
	config := ctx.Value("config").(*utils.Config)
	wg := ctx.Value("wg").(*sync.WaitGroup)

	defer wg.Done()

	logger := utils.NewLogger(config.DbName, "worker")
	ctx = context.WithValue(ctx, "logger", logger)

	logger.Info("starting worker")
	logger.With("count", unsentPostsCount(ctx)).With("wait", config.Interval).Info("unsent posts remaining")

	bot, err := telegram_bot.CreateBot(ctx, logger)
	if err != nil {
		logger.With("err", err).Error("failed to create bot")
		return
	}

	ticker := time.NewTimer(config.Interval)
	for {
		select {
		case <-ticker.C:
			postObj, err := db.Post.Query().
				Where(post.IsSent(false)).
				Order(sql.OrderByRand()).
				First(ctx)
			if ent.IsNotFound(err) {
				logger.With("wait", config.Interval).Info("no unsent posts found")
				continue
			} else if err != nil {
				logger.With("err", err).Error("error fetching posts")
				continue
			}
			_ = sendPost(postObj, bot, ctx)
			logger.With("count", unsentPostsCount(ctx)).
				With("wait", config.Interval).
				Info("unsent posts remaining")

		case <-ctx.Done():
			return
		}
	}
}
