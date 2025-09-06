package handlers

import (
	"channel-helper-go/database"
	"channel-helper-go/utils"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GifHandlerBasePayload struct {
	Base64 string `json:"base64" binding:"required"`
}

type GifHandlerUrlPayload struct {
	Url string `json:"url" binding:"required"`
}

func GifHandler(c *gin.Context) {
	db := c.MustGet("db").(*database.DBStruct)
	hub := c.MustGet("hub").(*utils.Hub)
	logger := c.MustGet("logger").(utils.Logger)

	var err error
	var gifBytes []byte
	if fileHeader, err := c.FormFile("upload"); err == nil {
		file, _ := fileHeader.Open()
		defer func(file multipart.File) {
			_ = file.Close()
		}(file)
		gifBytes = make([]byte, fileHeader.Size)
		_, err = file.Read(gifBytes)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Failed to read file"})
			return
		}
	} else if base64Data := c.PostForm("base64"); base64Data != "" {
		gifBytes, err = base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "Invalid base64 data"})
			return
		}
	} else {
		var payloadBase64 GifHandlerBasePayload
		var payloadUrl GifHandlerUrlPayload
		if err = c.ShouldBindBodyWithJSON(&payloadBase64); err == nil {
			gifBytes, err = base64.StdEncoding.DecodeString(payloadBase64.Base64)
			if err != nil {
				c.JSON(400, gin.H{"status": "error", "message": "Invalid base64 data"})
				return
			}
		} else if err = c.ShouldBindBodyWithJSON(&payloadUrl); err == nil {
			resp, err := http.Get(payloadUrl.Url)
			if err != nil {
				c.JSON(400, gin.H{"status": "error", "message": "Failed to fetch image from URL"})
				return
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if gifBytes, err = io.ReadAll(resp.Body); err != nil && err != io.EOF {
				c.JSON(500, gin.H{"status": "error", "message": "Failed to read image from URL"})
				return
			}
		} else {
			c.JSON(400, gin.H{"status": "error", "message": "No image data provided"})
			return
		}
	}

	if gifBytes, err = utils.ConvertToMp4(c, gifBytes); err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to convert GIF to MP4"})
		logger.With("err", err).Error("Failed to convert GIF to MP4")
		return
	}

	task := &database.UploadTask{
		Type: database.MediaTypeAnimation,
		Data: &gifBytes,
	}
	err = db.UploadTask.Create(c, task)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to create upload task"})
		return
	}

	hub.UploadTaskCreated <- task

	c.JSON(200, gin.H{"status": "ok", "upload_id": task.ID})
}
