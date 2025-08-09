package handlers

import (
	"channel-helper-go/ent"
	"channel-helper-go/ent/uploadtask"
	"channel-helper-go/utils"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
)

type GifHandlerBasePayload struct {
	Base64 string `json:"base64" binding:"required"`
}

type GifHandlerUrlPayload struct {
	Url string `json:"url" binding:"required"`
}

func GifHandler(c *gin.Context) {
	db := c.MustGet("db").(*ent.Client)
	hub := c.MustGet("hub").(*utils.Hub)

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
		var payloadBase64 PhotoHandlerBasePayload
		var payloadUrl PhotoHandlerUrlPayload
		if err = c.BindJSON(&payloadBase64); err == nil {
			gifBytes, err = base64.StdEncoding.DecodeString(payloadBase64.Base64)
			if err != nil {
				c.JSON(400, gin.H{"status": "error", "message": "Invalid base64 data"})
				return
			}
		} else if err = c.BindJSON(&payloadUrl); err == nil {
			resp, err := http.Get(payloadUrl.Url)
			if err != nil {
				c.JSON(400, gin.H{"status": "error", "message": "Failed to fetch image from URL"})
				return
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			gifBytes = make([]byte, resp.ContentLength)
			_, err = resp.Body.Read(gifBytes)
			if err != nil && err != io.EOF {
				c.JSON(500, gin.H{"status": "error", "message": "Failed to read image from URL"})
				return
			}
		} else {
			c.JSON(400, gin.H{"status": "error", "message": "No image data provided"})
			return
		}
	}

	uploadTask, err := db.UploadTask.Create().
		SetType(uploadtask.TypeAnimation).
		SetData(gifBytes).
		Save(c)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to create upload task"})
		return
	}

	hub.UploadTaskCreated <- uploadTask

	c.JSON(200, gin.H{"status": "ok", "upload_id": uploadTask.ID})
}
