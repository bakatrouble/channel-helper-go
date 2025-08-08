package web_api

import (
	ent "channel-helper-go/ent"
	channels "channel-helper-go/modules"
	"channel-helper-go/modules/web-api/handlers"
	"channel-helper-go/utils"
	"context"
	"fmt"
	"github.com/gin-contrib/graceful"
	"github.com/gin-gonic/gin"
	"sync"
)

func StartWebAPI(ctx context.Context) {
	config := ctx.Value("config").(*utils.Config)
	db := ctx.Value("db").(*ent.Client)
	wg := ctx.Value("wg").(*sync.WaitGroup)
	hub := ctx.Value("hub").(*channels.Hub)

	defer wg.Done()

	router, _ := graceful.Default(
		graceful.WithAddr(fmt.Sprintf("127.0.0.1:%d", config.ApiPort)),
	)
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

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
		c.Next()
	})

	g.POST("/photo", handlers.PhotoHandler)
	g.POST("/gif", handlers.GifHandler)
	g.GET("/hashes", handlers.HashesHandler)
	g.GET("/ws", handlers.WebsocketHandler)

	_ = router.RunWithContext(ctx)
}
