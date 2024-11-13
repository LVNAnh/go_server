package Controllers

import (
	"context"
	"net/http"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	serviceCollection := getServiceCollection()
	var service Models.Service
	if err := serviceCollection.FindOne(context.Background(), bson.M{"_id": orderBookingService.ServiceID}).Decode(&service); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	orderBookingService.UserID = userID
	orderBookingService.TotalPrice = float64(orderBookingService.Quantity) * service.Price
	orderBookingService.Status = "pending"
	orderBookingService.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	orderBookingService.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	orderBookingService.BookingDate = primitive.NewDateTimeFromTime(time.Now())

	orderBookingServiceCollection := getOrderBookingServiceCollection()
	if _, err := orderBookingServiceCollection.InsertOne(context.Background(), orderBookingService); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order booking service"})
		return
	}

	c.JSON(http.StatusOK, orderBookingService)
}

func GetOrderBookingServices(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	orderBookingCollection := getOrderBookingServiceCollection()
	var orderBookings []Models.OrderBookingService
	cursor, err := orderBookingCollection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order bookings"})
		return
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &orderBookings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode order bookings"})
		return
	}

	c.JSON(http.StatusOK, orderBookings)
}

func GetAllOrderBookingServices(c *gin.Context) {
	orderBookingCollection := getOrderBookingServiceCollection()
	var orderBookings []Models.OrderBookingService

	cursor, err := orderBookingCollection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get all order bookings"})
		return
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &orderBookings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode order bookings"})
		return
	}

	c.JSON(http.StatusOK, orderBookings)
}

func UpdateOrderBookingServiceStatus(c *gin.Context) {
	orderID := c.Param("id")

	var statusUpdate struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if statusUpdate.Status != "pending" && statusUpdate.Status != "confirmed" &&
		statusUpdate.Status != "in-progress" && statusUpdate.Status != "completed" &&
		statusUpdate.Status != "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	orderBookingServiceCollection := getOrderBookingServiceCollection()
	orderIDObj, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var orderBookingService Models.OrderBookingService
	if err := orderBookingServiceCollection.FindOne(context.Background(), bson.M{"_id": orderIDObj}).Decode(&orderBookingService); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if orderBookingService.Status == "completed" || orderBookingService.Status == "cancelled" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update a completed or cancelled order"})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"status":     statusUpdate.Status,
			"updated_at": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	if _, err := orderBookingServiceCollection.UpdateOne(context.Background(), bson.M{"_id": orderIDObj}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated"})
}
