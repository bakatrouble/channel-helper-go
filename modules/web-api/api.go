package web_api

import (
	"channel-helper-go/database"
	"channel-helper-go/modules/web-api/handlers"
	"channel-helper-go/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-contrib/graceful"
	"github.com/gin-gonic/gin"
	"github.com/samber/slog-gin"
	"sync"
)

func StartWebAPI(ctx context.Context) {
	config := ctx.Value("config").(*utils.Config)
	wg := ctx.Value("wg").(*sync.WaitGroup)
	sqldb := ctx.Value("sqldb").(*sql.DB)
	hub := ctx.Value("hub").(*utils.Hub)

	logger := utils.NewLogger(config.DbName, "web-api")
	logger.With("host", "127.0.0.1").With("port", config.ApiPort).Info("starting web api")

	defer wg.Done()

	db, err := database.NewDBStruct(sqldb, !config.Production, logger)
	if err != nil {
		logger.With("err", err).Error("failed to connect to database")
		panic(err)
	}
	defer func(db *database.DBStruct) {
		_ = db.Close()
	}(db)

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		logger.With("method", httpMethod).With("path", absolutePath).With("handlers", nuHandlers).Debug("registered route")
	}
	gin.DebugPrintFunc = func(format string, v ...any) {
		logger.Debug(format, v...)
	}
	if config.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	router, _ := graceful.Default(
		graceful.WithAddr(fmt.Sprintf("127.0.0.1:%d", config.ApiPort)),
	)
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	router.Use(sloggin.New(logger))
	router.Use(gin.Recovery())

	g := router.Group("/:api_key")

	g.Use(func(c *gin.Context) {
		apiKey := c.Param("api_key")
		if apiKey != config.ApiKey {
			c.JSON(403, gin.H{"status": "error", "message": "Forbidden"})
			c.Abort()
			return
		}
		c.Set("config", config)
		c.Set("db", db)
		c.Set("hub", hub)
		c.Set("logger", logger)
		c.Next()
	})

	g.POST("/photo", handlers.PhotoHandler)
	g.POST("/gif", handlers.GifHandler)
	g.GET("/hashes", handlers.HashesHandler)
	g.GET("/ws", handlers.WebsocketHandler)

	err = router.RunWithContext(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.With("err", err).Error("failed to start web api")
	}
}
