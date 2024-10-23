package Models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role int

const (
	Admin Role = iota
	Staff
	Customer
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FirstName string             `json:"firstname,omitempty"`
	LastName  string             `json:"lastname,omitempty"`
	Email     string             `json:"email,omitempty"`
	Password  string             `json:"password,omitempty"`
	Phone     string             `json:"phone,omitempty"`
	Address   string             `json:"address,omitempty"`
	Role      Role               `json:"role,omitempty"`
	Avatar    string             `json:"avatar,omitempty"`
	Cart      Cart               `json:"cart,omitempty"`
}

type Cart struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Items     []CartItem         `bson:"items,omitempty" json:"items,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type CartItem struct {
	ProductID primitive.ObjectID `bson:"product_id,omitempty" json:"product_id,omitempty"`
	Quantity  int                `bson:"quantity,omitempty" json:"quantity,omitempty"`
	Price     float64            `bson:"price,omitempty" json:"price,omitempty"`
	Name      string             `bson:"-" json:"name,omitempty"`
	ImageURL  string             `bson:"-" json:"imageurl,omitempty"`
}

type SelectedItems struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Items     []SelectedItem     `bson:"items,omitempty" json:"items,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type SelectedItem struct {
	ProductID primitive.ObjectID `bson:"product_id,omitempty" json:"product_id,omitempty"`
	Quantity  int                `bson:"quantity,omitempty" json:"quantity,omitempty"`
	Price     float64            `bson:"price,omitempty" json:"price,omitempty"`
	Name      string             `bson:"name,omitempty" json:"name,omitempty"`
	ImageURL  string             `bson:"imageurl,omitempty" json:"imageurl,omitempty"`
}

type Order struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID     primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Items      []OrderItem        `bson:"items,omitempty" json:"items,omitempty"`
	TotalPrice float64            `bson:"total_price,omitempty" json:"total_price,omitempty"`
	CreatedAt  time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt  time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type OrderItem struct {
	ProductID primitive.ObjectID `bson:"product_id,omitempty" json:"product_id,omitempty"`
	Quantity  int                `bson:"quantity,omitempty" json:"quantity,omitempty"`
	Price     float64            `bson:"price,omitempty" json:"price,omitempty"`
	Name      string             `bson:"name,omitempty" json:"name,omitempty"`
	ImageURL  string             `bson:"imageurl,omitempty" json:"imageurl,omitempty"`
}

type OrderBookingService struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	ServiceID    primitive.ObjectID `bson:"service_id" json:"service_id"`
	Quantity     int                `bson:"quantity" json:"quantity"`
	TotalPrice   float64            `bson:"total_price" json:"total_price"`
	BookingDate  primitive.DateTime `bson:"booking_date" json:"booking_date"`
	ContactName  string             `bson:"contact_name" json:"contact_name"`
	ContactPhone string             `bson:"contact_phone" json:"contact_phone"`
	Address      string             `bson:"address" json:"address"`
	Status       string             `bson:"status" json:"status"`
	CreatedAt    primitive.DateTime `bson:"created_at" json:"created_at"`
	UpdatedAt    primitive.DateTime `bson:"updated_at" json:"updated_at"`
	FinishAt     primitive.DateTime `bson:"finish_at" json:"finish_at"`
	Note         string             `bson:"note" json:"note"`
}
