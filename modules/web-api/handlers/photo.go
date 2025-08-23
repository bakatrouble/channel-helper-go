package handlers

import (
	"bytes"
	"channel-helper-go/database"
	"channel-helper-go/utils"
	"encoding/base64"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

type PhotoHandlerBasePayload struct {
	Base64 string `json:"base64" binding:"required"`
}

type PhotoHandlerUrlPayload struct {
	Url string `json:"url" binding:"required"`
}

func PhotoHandler(c *gin.Context) {
	db := c.Value("db").(*database.DBStruct)
	hub := c.MustGet("hub").(*utils.Hub)
	logger := c.MustGet("logger").(utils.Logger)

	var imageBytes []byte
	var err error
	if fileHeader, err := c.FormFile("upload"); err == nil {
		file, _ := fileHeader.Open()
		defer func(file multipart.File) {
			_ = file.Close()
		}(file)
		imageBytes = make([]byte, fileHeader.Size)
		if _, err = file.Read(imageBytes); err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Failed to read file"})
			logger.With("err", err).Error("failed to read file")
			return
		}
	} else if base64Data := c.PostForm("base64"); base64Data != "" {
		imageBytes, err = base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "Invalid base64 data"})
			logger.With("err", err).Error("invalid base64 data")
			return
		}
	} else {
		var payloadBase64 PhotoHandlerBasePayload
		var payloadUrl PhotoHandlerUrlPayload
		if err = c.ShouldBindBodyWithJSON(&payloadBase64); err == nil {
			imageBytes, err = base64.StdEncoding.DecodeString(payloadBase64.Base64)
			if err != nil {
				c.JSON(400, gin.H{"status": "error", "message": "Invalid base64 data"})
				logger.With("err", err).Error("invalid base64 data")
				return
			}
		} else if err = c.ShouldBindBodyWithJSON(&payloadUrl); err == nil {
			resp, err := http.Get(payloadUrl.Url)
			if err != nil {
				c.JSON(400, gin.H{"status": "error", "message": "Failed to fetch image from URL"})
				logger.With("err", err).With("url", payloadUrl.Url).Error("failed to fetch image from url")
				return
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			imageBytes = make([]byte, resp.ContentLength)

			if _, err = resp.Body.Read(imageBytes); err != nil && err != io.EOF {
				c.JSON(500, gin.H{"status": "error", "message": "Failed to read image from URL"})
				logger.With("err", err).With("url", payloadUrl.Url).Error("failed to read image from url")
				return
			}
		} else {
			c.JSON(400, gin.H{"status": "error", "message": "No image data provided"})
			logger.Error("no image data provided")
			return
		}
	}

	imConfig, _, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid image format"})
		logger.With("err", err).Error("invalid image format")
		return
	}
	im, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid image format"})
		logger.With("err", err).Error("invalid image format")
		return
	}
	if imConfig.Width > 2000 || imConfig.Height > 2000 {
		im = resize.Resize(2000, 2000, im, resize.Lanczos3)
	}
	imageBuffer := new(bytes.Buffer)
	if err = jpeg.Encode(imageBuffer, im, &jpeg.Options{Quality: 100}); err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to encode image"})
		logger.With("err", err).Error("failed to encode image")
		return
	}
	imageBytes = imageBuffer.Bytes()

	hash, err := utils.HashImage(imageBytes)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to hash image"})
		logger.With("err", err).Error("failed to hash image")
		return
	}
	logger.With("hash", hash).Info("image hash calculated")

	duplicate, _, _, err := db.ImageHash.Exists(c, hash)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Database error"})
		logger.With("err", err).Error("error checking for duplicate photo hash")
		return
	}
	if duplicate {
		c.JSON(400, gin.H{"status": "duplicate", "hash": hash})
		logger.With("hash", hash).Info("duplicate photo hash found")
		return
	}

	task := &database.UploadTask{
		Type: database.MediaTypePhoto,
		Data: &imageBytes,
		ImageHash: &database.ImageHash{
			Hash: hash,
		},
	}
	if err = db.UploadTask.Create(c, task); err != nil {
		c.JSON(500, gin.H{"status": "error", "message": "Failed to create upload task"})
		logger.With("err", err).Error("failed to create upload task")
		return
	}

	hub.UploadTaskCreated <- task

	c.JSON(200, gin.H{"status": "ok", "hash": hash, "upload_id": task.ID})
}
