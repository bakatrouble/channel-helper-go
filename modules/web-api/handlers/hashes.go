package handlers

import (
	"channel-helper-go/database"
	"channel-helper-go/utils"
	"github.com/gin-gonic/gin"
)

func HashesHandler(c *gin.Context) {
	db := c.MustGet("db").(*database.DBStruct)
	logger := c.MustGet("logger").(utils.Logger)

	hashes, err := db.ImageHash.GetAll(c)
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
	db := c.MustGet("db").(*database.DBStruct)
	logger := c.MustGet("logger").(utils.Logger)

	hash := c.Param("hash")
	if hash == "" {
		c.JSON(400, gin.H{"status": "error", "message": "Hash parameter is required"})
		return
	}

	exists, _, _, err := db.ImageHash.Exists(c, hash)
	if err != nil {
		logger.With("err", err).Error("failed to check if hash exists")
		c.JSON(500, gin.H{"status": "error", "message": "Failed to check hash existence"})
		return
	}

	c.JSON(200, gin.H{"status": "success", "exists": exists})
}
