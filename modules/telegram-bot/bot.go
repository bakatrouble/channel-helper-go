package telegram_bot

import (
	"channel-helper-go/ent"
	"channel-helper-go/modules/telegram-bot/handlers"
	"channel-helper-go/utils"
	"context"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"slices"
	"strings"
	"sync"
	"time"
)

type botLogger struct {
	utils.Logger
	replacer *strings.Replacer
}

func (b *botLogger) Debugf(format string, args ...interface{}) {
	b.Debug(b.replacer.Replace(fmt.Sprintf(format, args...)))
}

func (b *botLogger) Errorf(format string, args ...interface{}) {
	b.Error(b.replacer.Replace(fmt.Sprintf(format, args...)))
}

func CreateBot(ctx context.Context, logger utils.Logger) (*telego.Bot, error) {
	config := ctx.Value("config").(*utils.Config)

	l := botLogger{logger, strings.NewReplacer(config.BotToken, "****")}

	bot, err := telego.NewBot(
		config.BotToken,
		telego.WithLogger(&l),
	)
	return bot, err
}

func StartBot(ctx context.Context) {
	config := ctx.Value("config").(*utils.Config)
	db := ctx.Value("db").(*ent.Client)
	wg := ctx.Value("wg").(*sync.WaitGroup)
	hub := ctx.Value("hub").(*utils.Hub)

	logger := utils.NewLogger(config.DbName, "telegram-bot")
	logger.Info("starting telegram bot")

	defer wg.Done()

	bot, err := CreateBot(ctx, logger)
	if err != nil {
		logger.With("err", err).Error("failed to create bot")
		return
	}

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)
	bh, _ := th.NewBotHandler(bot, updates)
	defer func() {
		_ = bh.Stop()
	}()

	bh.Use(func(ctx *th.Context, update telego.Update) error {
		ctx = ctx.WithValue("config", config)
		ctx = ctx.WithValue("db", db)
		ctx = ctx.WithValue("wg", wg)
		ctx = ctx.WithValue("hub", hub)
		ctx = ctx.WithValue("logger", logger)
		return ctx.Next(update)
	})
	bh.Use(func(ctx *th.Context, update telego.Update) error {
		if update.Message == nil {
			return ctx.Next(update)
		}

		if update.Message.From == nil {
			return ctx.Next(update)
		}

		if slices.Contains(config.AllowedSenderChats, update.Message.Chat.ID) {
			return ctx.Next(update)
		}

		_, _ = bot.SendMessage(ctx, tu.Message(
			update.Message.Chat.ChatID(),
			"GTFO",
		))
		return nil
	})
	bh.HandleMessage(handlers.PhotoHandler, messageWithPhoto)
	bh.HandleMessage(handlers.AnimationHandler, messageWithAnimation)
	bh.HandleMessage(handlers.VideoHandler, messageWithVideo)
	bh.HandleMessage(handlers.DeleteCommandHandler, messageCommands([]string{"delete", "del", "remove", "rem", "rm"}))
	bh.HandleMessage(handlers.CountHandler, messageCommands([]string{"count", "cnt"}))
	bh.HandleMessage(handlers.DumpDbHandler, messageCommands([]string{"dump", "dumpdb", "dump_db"}))
	bh.HandleMessage(handlers.UnknownHandler, th.AnyMessage())
	bh.HandleCallbackQuery(handlers.DeleteCallbackHandler, th.CallbackDataEqual("/delete"))

	// Initialize done chan
	done := make(chan struct{}, 1)

	// Handle stop signal (Ctrl+C)
	go func() {
		<-ctx.Done()
		logger.Warn("stopping telegram bot")
		stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second*20)
		defer stopCancel()

	out:
		for len(updates) > 0 {
			select {
			case <-stopCtx.Done():
				break out
			case <-time.After(time.Millisecond * 100):
				// Continue
			}
		}
		logger.Info("long polling done")
		_ = bh.StopWithContext(stopCtx)
		done <- struct{}{}
	}()

	go func() { _ = bh.Start() }()
	logger.Info("handling updates")

	<-done
	logger.Info("telegram bot stopped")
}

func messageWithAnimation(_ context.Context, update telego.Update) bool {
	return update.Message != nil && update.Message.Animation != nil
}

func messageWithVideo(_ context.Context, update telego.Update) bool {
	return update.Message != nil && update.Message.Video != nil
}

func messageWithPhoto(_ context.Context, update telego.Update) bool {
	return update.Message != nil && len(update.Message.Photo) > 0
}

func messageCommands(commands []string) th.Predicate {
	return func(_ context.Context, update telego.Update) bool {
		if update.Message == nil {
			return false
		}

		matches := th.CommandRegexp.FindStringSubmatch(update.Message.Text)
		if len(matches) != th.CommandMatchGroupsLen {
			return false
		}

		for _, command := range commands {
			if strings.EqualFold(matches[th.CommandMatchCmdGroup], command) {
				return true
			}
		}
		return false
	}
}
