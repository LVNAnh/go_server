package Controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"Server/Middleware"
	"Server/Models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

	if chat.CustomerID == primitive.NilObjectID && (chat.GuestName == "" || chat.GuestPhone == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guest name and phone required"})
		return
	}

	collection := Database.Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"$or": []bson.M{
		{"customer_id": chat.CustomerID, "is_active": true},
		{"guest_phone": chat.GuestPhone, "is_active": true},
	}}
	var existingChat Models.SupportChat
	err := collection.FindOne(ctx, filter).Decode(&existingChat)
	if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Active chat already exists"})
		return
	}

	chat.ID = primitive.NewObjectID()
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()
	chat.IsActive = true

	_, err = collection.InsertOne(ctx, chat)
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

	user, _ := c.Get("user")
	claims := user.(*Middleware.UserClaims)

	if claims.Role != 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can reply"})
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

func GetAllChatsAndMessages(c *gin.Context) {
	user, _ := c.Get("user")
	claims := user.(*Middleware.UserClaims)
	if claims.Role != 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	collection := Database.Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var chats []Models.SupportChat
	cursor, err := collection.Find(ctx, bson.M{"customer_id": primitive.NilObjectID, "is_active": true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching chats"})
		return
	}
	if err := cursor.All(ctx, &chats); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing chats"})
		return
	}

	c.JSON(http.StatusOK, chats)
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

	}
}
