package handlers

import (
	"bytes"
	"channel-helper-go/ent"
	"channel-helper-go/ent/uploadtask"
	"channel-helper-go/utils"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
)

type PhotoHandlerBasePayload struct {
	Base64 string `json:"base64" binding:"required"`
}

type PhotoHandlerUrlPayload struct {
	Url string `json:"url" binding:"required"`
}

func PhotoHandler(c *gin.Context) {
	db := c.Value("db").(*ent.Client)
	hub := c.MustGet("hub").(*utils.Hub)

	var imageBytes []byte
	if fileHeader, err := c.FormFile("upload"); err == nil {
		file, _ := fileHeader.Open()
		defer func(file multipart.File) {
			_ = file.Close()
		}(file)
		imageBytes = make([]byte, fileHeader.Size)
		_, err := file.Read(imageBytes)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Failed to read file"})
			return
		}
	} else if base64Data := c.PostForm("base64"); base64Data != "" {
		imageBytes, err = base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "Invalid base64 data"})
			return
		}
	} else {
		var payloadBase64 PhotoHandlerBasePayload
		var payloadUrl PhotoHandlerUrlPayload
		if err = c.BindJSON(&payloadBase64); err == nil {
			imageBytes, err = base64.StdEncoding.DecodeString(payloadBase64.Base64)
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
			imageBytes = make([]byte, resp.ContentLength)
			_, err = resp.Body.Read(imageBytes)
			if err != nil && err != io.EOF {
				c.JSON(500, gin.H{"status": "error", "message": "Failed to read image from URL"})
				return
			}
		} else {
			c.JSON(400, gin.H{"status": "error", "message": "No image data provided"})
			return
		}
	}

	imConfig, _, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid image format"})
		return
	}
	im, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid image format"})
		return
	}
	if imConfig.Width > 2000 || imConfig.Height > 2000 {
		im = resize.Resize(2000, 2000, im, resize.Lanczos3)
	}
	imageBytes = make([]byte, 0)
	err = jpeg.Encode(bytes.NewBuffer(imageBytes), im, &jpeg.Options{Quality: 85})
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to encode image"})
		return
	}

	hash, err := utils.HashImage(imageBytes)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to hash image"})
		return
	}

	duplicate, _, _, err := ent.PhotoHashExists(hash, c, db)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Database error"})
		return
	}
	if duplicate {
		c.JSON(400, gin.H{"status": "duplicate", "hash": hash})
		return
	}

	tx, err := db.Tx(c)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to start transaction"})
		return
	}
	uploadTask, err := tx.UploadTask.Create().
		SetType(uploadtask.TypePhoto).
		SetData(imageBytes).
		Save(c)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to create upload task"})
		return
	}
	err = tx.ImageHash.Create().
		SetImageHash(hash).
		SetUploadTask(uploadTask).
		Exec(c)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to create image hash"})
		return
	}
	if err = tx.Commit(); err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to commit transaction"})
		return
	}

	hub.UploadTaskCreated <- uploadTask

	c.JSON(200, gin.H{"status": "ok", "hash": hash, "upload_id": uploadTask.ID})
}
