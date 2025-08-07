package web_api

import (
	ent "channel-helper-go/ent"
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
	uploadTaskChannel := ctx.Value("uploadTaskChannel").(chan *ent.UploadTask)

	defer wg.Done()

	router, _ := graceful.Default(
		graceful.WithAddr(fmt.Sprintf("127.0.0.1:%d", config.ApiPort)),
	)
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	router.Use(func(c *gin.Context) {
		c.Set("config", config)
		c.Set("db", db)
		c.Set("uploadTaskChannel", uploadTaskChannel)
	})

	router.POST("/photo", handlers.PhotoHandler)
	router.POST("/gif", handlers.GifHandler)

	_ = router.RunWithContext(ctx)
}
