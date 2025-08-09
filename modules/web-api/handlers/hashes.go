package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"channel-helper-go/utils"
	"github.com/gin-gonic/gin"
)

func HashesHandler(c *gin.Context) {
	db := c.MustGet("db").(*ent.Client)
	logger := c.MustGet("logger").(utils.Logger)

	hashes, err := db.Post.Query().
		Where(post.ImageHashNotNil()).
		Select(post.FieldImageHash).
		Strings(c)
	if err != nil {
		logger.With("err", err).Error("failed to get hashes")
		c.JSON(500, gin.H{"status": "error", "message": "Failed to retrieve hashes"})
		return
	}

	if hashes == nil {
		hashes = []string{}
	}
	c.JSON(200, gin.H{"status": "success", "hashes": hashes})
}
