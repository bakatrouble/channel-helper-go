package scripts

import (
	"channel-helper-go/ent"
	telegrambot "channel-helper-go/modules/telegram-bot"
	"channel-helper-go/modules/uploader"
	webapi "channel-helper-go/modules/web-api"
	"channel-helper-go/modules/worker"
	"channel-helper-go/utils"
	"context"
	"fmt"
	"github.com/DrSmithFr/go-console"
	"github.com/mymmrac/telego"
	"os"
	"os/signal"
	"sync"
)

func StartScript(cmd *go_console.Script) go_console.ExitCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := utils.ParseConfig(cmd.Input.Option("config"))
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to parse config file: %v", err)
		return go_console.ExitError
	}
	ctx = context.WithValue(ctx, "config", config)

	db, err := ent.ConnectToDB(config.DbName, ctx)
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to connect to database: %v", err)
		return go_console.ExitError
	}
	defer func(db *ent.Client) {
		_ = db.Close()
	}(db)
	ctx = context.WithValue(ctx, "db", db)

	wg := sync.WaitGroup{}
	ctx = context.WithValue(ctx, "wg", &wg)

	bot, err := telego.NewBot(config.BotToken, telego.WithDefaultLogger(false, true))
	if err != nil {
		_, _ = fmt.Fprintf(cmd, "Failed to create bot: %v", err)
		return go_console.ExitError
	}
	ctx = context.WithValue(ctx, "bot", bot)

	uploadTaskChannel := make(chan *ent.UploadTask)
	ctx = context.WithValue(ctx, "uploadTaskChannel", uploadTaskChannel)
	postChannel := make(chan *ent.Post)
	ctx = context.WithValue(ctx, "postChannel", postChannel)

	go telegrambot.StartBot(ctx)
	wg.Add(1)
	go worker.StartWorker(ctx)
	wg.Add(1)

	if config.WithApi {
		go webapi.StartWebAPI(ctx)
		wg.Add(1)
		go uploader.StartUploader(ctx)
		wg.Add(1)
	}

	wg.Wait()

	return go_console.ExitSuccess
}
