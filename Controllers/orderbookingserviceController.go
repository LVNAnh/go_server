package Controllers

import (
	"context"
	"time"

	"Server/Middleware"
	"Server/Models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getOrderBookingServiceCollection() *mongo.Collection {
	return Database.Collection("order_booking_service")
}

func getServiceCollection() *mongo.Collection {
	return Database.Collection("services")
}

func CreateOrderBookingService(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	var orderBookingService Models.OrderBookingService
	if err := c.ShouldBindJSON(&orderBookingService); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	serviceCollection := getServiceCollection()
	var service Models.Service
	if err := serviceCollection.FindOne(context.Background(), bson.M{"_id": orderBookingService.ServiceID}).Decode(&service); err != nil {
		c.JSON(404, gin.H{"error": "Service not found"})
		return
	}

	orderBookingService.UserID = userID
	orderBookingService.TotalPrice = float64(orderBookingService.Quantity) * service.Price
	orderBookingService.Status = "Chờ xác nhận"
	orderBookingService.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	orderBookingService.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	orderBookingService.BookingDate = primitive.NewDateTimeFromTime(time.Now())

	orderBookingServiceCollection := getOrderBookingServiceCollection()
	if _, err := orderBookingServiceCollection.InsertOne(context.Background(), orderBookingService); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create order booking service"})
		return
	}

	c.JSON(200, orderBookingService)
}

func GetOrderBookingServices(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	orderBookingCollection := getOrderBookingServiceCollection()
	var orderBookings []Models.OrderBookingService
	cursor, err := orderBookingCollection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get order bookings"})
		return
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &orderBookings); err != nil {
		c.JSON(500, gin.H{"error": "Failed to decode order bookings"})
		return
	}

	c.JSON(200, orderBookings)
}

func UpdateOrderBookingServiceStatus(c *gin.Context) {
	orderID := c.Param("id")

	var statusUpdate struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	if statusUpdate.Status != "Chờ xác nhận" && statusUpdate.Status != "Đã xác nhận" &&
		statusUpdate.Status != "Đang tiến hành" && statusUpdate.Status != "Hoàn thành" &&
		statusUpdate.Status != "Đã hủy" {
		c.JSON(400, gin.H{"error": "Invalid status value"})
		return
	}

	orderBookingServiceCollection := getOrderBookingServiceCollection()
	orderIDObj, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid order ID"})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"status":     statusUpdate.Status,
			"updated_at": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	if _, err := orderBookingServiceCollection.UpdateOne(context.Background(), bson.M{"_id": orderIDObj}, update); err != nil {
		c.JSON(500, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(200, gin.H{"message": "Order status updated"})
}
