package handlers

import (
	"channel-helper-go/database"
	"channel-helper-go/utils"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vimeo/go-magic/magic"
)

type SharedChatPayload struct {
	Path string `json:"path"`
}

func InternalSendHandler(c *gin.Context) {
	db := c.MustGet("db").(*database.DBStruct)
	hub := c.MustGet("hub").(*utils.Hub)
	logger := c.MustGet("logger").(utils.Logger)

	var payload SharedChatPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid payload"})
		logger.With("err", err).Error("invalid payload")
		return
	}

	mediaBytes, err := os.ReadFile(payload.Path)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to read file"})
		logger.With("err", err).Error("failed to read file")
		return
	}

	mime := magic.MimeFromBytes(mediaBytes)

	var task *database.UploadTask
	switch mime {
	case "image/jpeg":
		hash, err := utils.HashImage(mediaBytes)
		if err != nil {
			logger.With("err", err).Error("error hashing image")
			c.JSON(500, gin.H{"status": "error", "message": "Failed to hash image"})
			return
		}
		duplicate, _, _, err := db.ImageHash.Exists(c, hash)
		if err != nil {
			logger.With("err", err).Error("error checking for duplicate image hash")
			c.JSON(500, gin.H{"status": "error", "message": "Failed to check for duplicate image"})
			return
		}
		if duplicate {
			logger.With("hash", hash).Info("duplicate photo hash found")
			c.JSON(200, gin.H{"status": "duplicate", "hash": hash})
			return
		}
		task = &database.UploadTask{
			Type: database.MediaTypePhoto,
			Data: &mediaBytes,
			ImageHash: &database.ImageHash{
				Hash: hash,
			},
		}
	case "video/mp4":
		var hasAudio bool
		var err error
		var t database.MediaType
		if hasAudio, err = utils.Mp4HasAudio(c, mediaBytes); err != nil {
			logger.With("err", err).Error("error checking mp4 for audio")
			c.JSON(500, gin.H{"status": "error", "message": "Failed to check mp4 for audio"})
			return
		} else if hasAudio {
			t = database.MediaTypeVideo
		} else {
			t = database.MediaTypeAnimation
		}
		task = &database.UploadTask{
			Type: t,
			Data: &mediaBytes,
		}
	default:
		c.JSON(400, gin.H{"status": "error", "message": "Unsupported file type: " + mime})
		logger.With("type", mime).Error("unsupported file type")
		return
	}

	if err = db.UploadTask.Create(c, task); err != nil {
		logger.With("err", err).Error("failed to create upload task")
		c.JSON(500, gin.H{"status": "error", "message": "Failed to create upload task"})
		return
	}

	hub.UploadTaskCreated <- task

	hash := ""
	if task.ImageHash != nil {
		hash = task.ImageHash.Hash
	}
	c.JSON(200, gin.H{"status": "ok", "hash": hash, "upload_id": task.ID})
}
