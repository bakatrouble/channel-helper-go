package web_api

import (
	"channel-helper-go/database"
	"channel-helper-go/modules/web-api/handlers"
	"channel-helper-go/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/graceful"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/samber/slog-gin"
	"github.com/telegram-mini-apps/init-data-golang"
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

	router.Use(static.Serve("/miniapp", static.LocalFile("./miniapp/dist", true)))

	g := router.Group("/:api_key")

	corsMiddleware := cors.Default()
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
		c.Set("cors", corsMiddleware)
		c.Next()
	})

	router.Use(corsMiddleware)
	g.OPTIONS("/*group", corsMiddleware)
	g.POST("/photo", corsMiddleware, handlers.PhotoHandler)
	g.POST("/gif", corsMiddleware, handlers.GifHandler)
	g.GET("/hashes", corsMiddleware, handlers.HashesHandler)
	g.GET("/ws", corsMiddleware, handlers.WebsocketHandler)
	g.GET("/count", corsMiddleware, func(c *gin.Context) {
		count, err := db.Post.UnsentCount(c)
		if err != nil {
			logger.With("err", err).Error("failed to get unsent count")
			c.JSON(500, gin.H{"status": "error", "message": "Internal Server Error"})
			return
		}
		c.JSON(200, gin.H{"status": "success", "count": count})
	})

	g2 := router.Group("")
	g2.Use(corsMiddleware)
	g2.GET("/apiKey", corsMiddleware, func(c *gin.Context) {
		// get init data from query params
		initData := c.Query("init_data")
		logger.With("init_data", initData).Info("got init data")
		expiration := time.Minute
		if !config.Production {
			expiration = 0
		}
		if err := initdata.Validate(initData, config.BotToken, expiration); err != nil {
			logger.With("init_data", initData).With("err", err).Error("invalid init data")
			c.JSON(400, gin.H{"status": "error", "message": "Invalid init data"})
			return
		}
		data, _ := initdata.Parse(initData)
		checkChat := data.Chat.ID
		if checkChat == 0 {
			checkChat = data.User.ID
		}
		if !slices.Contains(config.AllowedSenderChats, checkChat) {
			logger.With("chat_id", checkChat).Error("chat not allowed to access API")
			c.JSON(403, gin.H{"status": "error", "message": "Forbidden: chat not allowed"})
			return
		}
		c.JSON(200, gin.H{"apiKey": config.ApiKey})
	})

	err = router.RunWithContext(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.With("err", err).Error("failed to start web api")
	}
}
