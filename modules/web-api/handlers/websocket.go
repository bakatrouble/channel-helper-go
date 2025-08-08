package handlers

import (
	channels "channel-helper-go/modules"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WebsocketHandler(c *gin.Context) {
	chans := c.MustGet("chans").(*channels.AppChannels)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		println("Failed to upgrade connection:", err.Error())
		c.JSON(500, gin.H{"status": "error", "message": "Failed to upgrade connection"})
		return
	}
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(conn)

	for {
		select {
		case post := <-chans.PostCreated:
			_ = conn.WriteJSON(gin.H{"type": "postCreated", "post": post})
		case post := <-chans.PostSent:
			_ = conn.WriteJSON(gin.H{"type": "postSent", "post": post})
		case post := <-chans.PostDeleted:
			_ = conn.WriteJSON(gin.H{"type": "postDeleted", "post": post})
		case uploadTask := <-chans.UploadTaskCreated:
			_ = conn.WriteJSON(gin.H{"type": "uploadTaskCreated", "uploadTask": uploadTask})
		case uploadTask := <-chans.UploadTaskDone:
			_ = conn.WriteJSON(gin.H{"type": "uploadTaskDone", "uploadTask": uploadTask})
		}
	}

	return
}
