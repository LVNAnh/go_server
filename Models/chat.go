package Models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SupportChat struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty"`
	GuestID    string             `bson:"guest_id,omitempty" json:"guest_id,omitempty"`
	GuestName  string             `bson:"guest_name,omitempty" json:"guest_name,omitempty"`
	GuestPhone string             `bson:"guest_phone,omitempty" json:"guest_phone,omitempty"`
	AdminID    primitive.ObjectID `bson:"admin_id" json:"admin_id"`
	Messages   []Message          `bson:"messages" json:"messages"`
	IsActive   bool               `bson:"is_active" json:"is_active"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type Message struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChatID     primitive.ObjectID `bson:"chat_id" json:"chat_id"`
	SenderID   primitive.ObjectID `bson:"sender_id,omitempty" json:"sender_id,omitempty"`
	GuestName  string             `bson:"guest_name,omitempty" json:"guest_name,omitempty"`
	SenderRole string             `bson:"sender_role" json:"sender_role"`
	Content    string             `bson:"content" json:"content"`
	Timestamp  time.Time          `bson:"timestamp" json:"timestamp"`
	Seen       bool               `bson:"seen" json:"seen"`
}

type ChatNotification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	ChatID      primitive.ObjectID `bson:"chat_id" json:"chat_id"`
	UnreadCount int                `bson:"unread_count" json:"unread_count"`
	LastMessage string             `bson:"last_message" json:"last_message"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
