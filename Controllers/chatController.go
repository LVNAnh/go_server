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

var clients = make(map[*websocket.Conn]string)
var broadcast = make(chan Models.Message)

func CreateChat(c *gin.Context) {
	var chat Models.SupportChat
	if err := c.ShouldBindJSON(&chat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Kiểm tra khách hàng chưa đăng ký (Guest) có thông tin tên và số điện thoại
	if chat.CustomerID == primitive.NilObjectID && (chat.GuestName == "" || chat.GuestPhone == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guest name and phone are required"})
		return
	}

	collection := Database.Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tìm nếu đã có chat hoạt động cho khách hàng hoặc Guest này
	filter := bson.M{"$or": []bson.M{
		{"customer_id": chat.CustomerID, "is_active": true},
		{"guest_phone": chat.GuestPhone, "is_active": true},
	}}
	var existingChat Models.SupportChat
	err := collection.FindOne(ctx, filter).Decode(&existingChat)
	if err == nil {
		// Nếu đã có chat, trả về chat đó
		c.JSON(http.StatusOK, existingChat)
		return
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing chat"})
		return
	}

	// Tạo chat mới nếu chưa có
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

	chatId := c.Query("chatId")
	role := c.Query("role") // Lấy vai trò (Admin hoặc Guest)

	// Kết nối WebSocket vào chat với chatId và vai trò
	clients[conn] = role

	if role == "Admin" {
		// Cập nhật admin_id trong document của chat trong MongoDB
		collection := Database.Collection("chats")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		chatObjectID, _ := primitive.ObjectIDFromHex(chatId)
		update := bson.M{"$set": bson.M{"admin_id": chatObjectID}}
		_, err := collection.UpdateOne(ctx, bson.M{"_id": chatObjectID}, update)
		if err != nil {
			log.Println("Error updating admin_id:", err)
			return
		}
	}

	// Khởi động việc xử lý tin nhắn giữa Guest và Admin
	go handleMessages()

	for {
		var msg Models.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, conn)
			break
		}

		msg.Timestamp = time.Now()
		broadcast <- msg
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

	filter := bson.M{"is_active": true, "admin_id": primitive.NilObjectID}
	var chatRequests []Models.SupportChat
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching chat requests"})
		return
	}
	if err := cursor.All(ctx, &chatRequests); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding chat requests"})
		return
	}

	c.JSON(http.StatusOK, chatRequests)
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
