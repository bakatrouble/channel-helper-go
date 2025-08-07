package telegram_bot

import (
	"channel-helper-go/ent"
	"channel-helper-go/modules/telegram-bot/handlers"
	"channel-helper-go/utils"
	"context"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"sync"
)

func StartBot(ctx context.Context) {
	config := ctx.Value("config").(*utils.Config)
	db := ctx.Value("db").(*ent.Client)
	wg := ctx.Value("wg").(*sync.WaitGroup)
	bot := ctx.Value("bot").(*telego.Bot)

	defer wg.Done()

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)
	bh, _ := th.NewBotHandler(bot, updates)
	defer func() {
		_ = bh.Stop()
	}()

	bh.Use(func(ctx *th.Context, update telego.Update) error {
		ctx = ctx.WithValue("config", config)
		ctx = ctx.WithValue("db", db)
		ctx = ctx.WithValue("wg", wg)
		return ctx.Next(update)
	})
	bh.HandleMessage(handlers.PhotoHandler, messageWithPhoto)
	bh.HandleMessage(handlers.AnimationHandler, messageWithAnimation)
	bh.HandleMessage(handlers.VideoHandler, messageWithVideo)
	bh.HandleMessage(handlers.UnknownHandler, th.AnyMessage())

	_ = bh.Start()
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
