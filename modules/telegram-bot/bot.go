package telegram_bot

import (
	"channel-helper-go/ent"
	"channel-helper-go/modules/telegram-bot/handlers"
	"channel-helper-go/utils"
	"context"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"slices"
	"strings"
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
	bh.HandleMessage(handlers.DeleteHandler, messageCommands([]string{"delete", "del", "remove", "rem", "rm"}))
	bh.HandleMessage(handlers.UnknownHandler, th.AnyMessage())
	bh.HandleCallbackQuery(handlers.DeleteCallbackHandler, th.CallbackDataEqual("/delete"))

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
