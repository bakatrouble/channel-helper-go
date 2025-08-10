package scripts

import (
	"channel-helper-go/database"
	telegrambot "channel-helper-go/modules/telegram-bot"
	"channel-helper-go/modules/uploader"
	webapi "channel-helper-go/modules/web-api"
	"channel-helper-go/modules/worker"
	"channel-helper-go/utils"
	"context"
	"github.com/DrSmithFr/go-console"
	"os"
	"os/signal"
	"sync"
)

func StartScript(cmd *go_console.Script) go_console.ExitCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := utils.ParseConfig(cmd.Input.Option("config"))
	if err != nil {
		panic("Failed to parse config file: " + err.Error())
	}
	ctx = context.WithValue(ctx, "config", config)

	wg := sync.WaitGroup{}
	ctx = context.WithValue(ctx, "wg", &wg)

	hub := utils.NewHub()
	ctx = context.WithValue(ctx, "hub", &hub)

	sqldb, err := database.NewSQLDB(config.DbName)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	ctx = context.WithValue(ctx, "sqldb", sqldb)

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
