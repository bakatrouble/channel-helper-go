package handlers

import (
	"channel-helper-go/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/grbit/go-json"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WebsocketHandler(c *gin.Context) {
	//hub := c.MustGet("hub").(*utils.Hub)
	logger := c.MustGet("logger").(utils.Logger)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.With("err", err).Error("failed to upgrade connection")
		c.JSON(500, gin.H{"status": "error", "message": "Failed to upgrade connection"})
		return
	}
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(conn)

	msgIn := make(chan interface{})

	go func(conn *websocket.Conn, channel chan interface{}) {
		for {
			var msg interface{}
			err := conn.ReadJSON(&msg)
			if websocket.IsUnexpectedCloseError(err) {
				println("socket closed:", err.Error())
				return
			} else if err == nil {
				channel <- msg
			}
		}
	}(conn, msgIn)

	for {
		select {
		//case post := <-hub.PostCreated:
		//	_ = conn.WriteJSON(gin.H{"type": "postCreated", "post": post})
		//case post := <-hub.PostSent:
		//	_ = conn.WriteJSON(gin.H{"type": "postSent", "post": post})
		//case post := <-hub.PostDeleted:
		//	_ = conn.WriteJSON(gin.H{"type": "postDeleted", "post": post})
		//case uploadTask := <-hub.UploadTaskCreated:
		//	_ = conn.WriteJSON(gin.H{"type": "uploadTaskCreated", "uploadTask": uploadTask})
		//case uploadTask := <-hub.UploadTaskDone:
		//	_ = conn.WriteJSON(gin.H{"type": "uploadTaskDone", "uploadTask": uploadTask})
		case msg := <-msgIn:
			j, _ := json.Marshal(msg)
			logger.With("msg", j).Info("received message from websocket")
		case <-c.Done():
			return
		}
	}
}
