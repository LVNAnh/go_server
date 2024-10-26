package Controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"Server/Models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func CreateChat(c *gin.Context) {
	var chat Models.SupportChat
	if err := c.ShouldBindJSON(&chat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	chat.ID = primitive.NewObjectID()
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()
	chat.IsActive = true

	collection := Database.Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, chat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating chat"})
		return
	}

	c.JSON(http.StatusOK, chat)
}

func ReplyChat(c *gin.Context) {
	var msg Models.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	msg.ID = primitive.NewObjectID()
	msg.Timestamp = time.Now()

	collection := Database.Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending message"})
		return
	}

	c.JSON(http.StatusOK, msg)
}

func ChatWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to set websocket upgrade: ", err)
		return
	}
	defer conn.Close()

	for {
		var msg Models.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading json: ", err)
			break
		}

		// Handle the message (e.g., save to database, broadcast to other users)
		msg.ID = primitive.NewObjectID()
		msg.Timestamp = time.Now()

		collection := Database.Collection("messages")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = collection.InsertOne(ctx, msg)
		if err != nil {
			log.Println("Error inserting message: ", err)
			break
		}

		// Broadcast the message to other users (this part will depend on your implementation)
	}
}
