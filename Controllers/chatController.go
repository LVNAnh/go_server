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
	"go.mongodb.org/mongo-driver/mongo/options"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]string)
var broadcast = make(chan Models.Message)

func CreateChat(c *gin.Context) {
	var chat Models.SupportChat

	if err := c.ShouldBind(&chat); err != nil {
		log.Println("Error parsing request payload:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	log.Println("Received guest name:", chat.GuestName)
	log.Println("Received guest phone:", chat.GuestPhone)

	if chat.CustomerID == primitive.NilObjectID && (chat.GuestName == "" || chat.GuestPhone == "") {
		log.Println("Missing guest name or phone")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guest name and phone are required"})
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
	if err == nil {
		log.Println("Active chat already exists for this guest or customer")
		c.JSON(http.StatusBadRequest, gin.H{"error": "An active chat already exists for this customer or guest"})
		return
	} else if err != mongo.ErrNoDocuments {
		log.Println("Error checking existing chat:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing chat"})
		return
	}

	chat.ID = primitive.NewObjectID()
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()
	chat.IsActive = true

	_, err = collection.InsertOne(ctx, chat)
	if err != nil {
		log.Println("Error inserting new chat:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating chat"})
		return
	}

	log.Println("New chat created with ID:", chat.ID)
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

	collection := Database.Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{"$push": bson.M{"messages": msg}}
	_, err := collection.UpdateOne(ctx, bson.M{"_id": msg.ChatID}, update)
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
		log.Println("Failed to set websocket upgrade:", err)
		return
	}
	defer conn.Close()

	role := c.Query("role") // Retrieve the role (Admin or Guest) from query parameters
	clients[conn] = role    // Map connection to role

	go handleMessages() // Start the message handler

	for {
		var msg Models.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, conn)
			break
		}

		msg.Timestamp = time.Now()
		broadcast <- msg // Send message to broadcast channel
	}
}

func handleMessages() {
	for {
		msg := <-broadcast // Get the latest message from the broadcast channel
		for client, role := range clients {
			if role != msg.SenderRole { // Send only to opposite role (Admin or Guest)
				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("WebSocket error: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}

func GetNewChatRequests(c *gin.Context) {
	collection := Database.Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var chats []Models.SupportChat
	options := options.Find().SetSort(bson.D{{"created_at", -1}})
	cursor, err := collection.Find(ctx, bson.M{"is_active": true, "customer_id": primitive.NilObjectID}, options)
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

func GetChatMessages(c *gin.Context) {
	chatId := c.Param("chatId")
	objectId, err := primitive.ObjectIDFromHex(chatId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	collection := Database.Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var messages []Models.Message
	cursor, err := collection.Find(ctx, bson.M{"chat_id": objectId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching messages"})
		return
	}
	if err := cursor.All(ctx, &messages); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
