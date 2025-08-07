package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/post"
	"github.com/gin-gonic/gin"
)

func HashesHandler(c *gin.Context) {
	db := c.MustGet("db").(*ent.Client)

	hashes, err := db.Post.Query().
		Where(post.ImageHashNotNil()).
		Select(post.FieldImageHash).
		Strings(c)
	if err != nil {
		println("Failed to get hashes:", err.Error())
		c.JSON(500, gin.H{"status": "error", "message": "Failed to retrieve hashes"})
		return
	}

	c.JSON(200, gin.H{"status": "success", "hashes": hashes})
}
