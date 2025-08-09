package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/imagehash"
	"channel-helper-go/utils"
	"github.com/gin-gonic/gin"
)

func HashesHandler(c *gin.Context) {
	db := c.MustGet("db").(*ent.Client)
	logger := c.MustGet("logger").(utils.Logger)

	hashes, err := db.ImageHash.Query().
		Select(imagehash.FieldImageHash).
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

func HashExistsHandler(c *gin.Context) {
	db := c.MustGet("db").(*ent.Client)
	logger := c.MustGet("logger").(utils.Logger)

	hash := c.Param("hash")
	if hash == "" {
		c.JSON(400, gin.H{"status": "error", "message": "Hash parameter is required"})
		return
	}

	exists, err := db.ImageHash.Query().
		Where(imagehash.ImageHashEQ(hash)).
		Exist(c)
	if err != nil {
		logger.With("err", err).Error("failed to check if hash exists")
		c.JSON(500, gin.H{"status": "error", "message": "Failed to check hash existence"})
		return
	}

	c.JSON(200, gin.H{"status": "success", "exists": exists})
}
