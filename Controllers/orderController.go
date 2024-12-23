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

func getOrderCollection() *mongo.Collection {
	return Database.Collection("product_order")
}

func getUserCollection() *mongo.Collection {
	return Database.Collection("users")
}

func CreateOrder(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	selectedItemsCollection := getSelectedItemsCollection()
	var selectedItems Models.SelectedItems
	err := selectedItemsCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&selectedItems)
	if err == mongo.ErrNoDocuments {
		c.JSON(404, gin.H{"error": "No selected items found"})
		return
	}

	productCollection := getProductCollection()
	var orderItems []Models.OrderItem
	totalPrice := 0.0

	for _, selectedItem := range selectedItems.Items {
		var product Models.Product
		err := productCollection.FindOne(context.Background(), bson.M{"_id": selectedItem.ProductID}).Decode(&product)
		if err != nil {
			c.JSON(404, gin.H{"error": "Product not found"})
			return
		}

		orderItem := Models.OrderItem{
			ProductID: selectedItem.ProductID,
			Quantity:  selectedItem.Quantity,
			Price:     product.Price,
			Name:      product.Name,
			ImageURL:  product.ImageURL,
		}

		orderItems = append(orderItems, orderItem)
		totalPrice += product.Price * float64(selectedItem.Quantity)
	}

	order := Models.Order{
		UserID:     userID,
		Items:      orderItems,
		TotalPrice: totalPrice,
		Status:     "pending",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	orderCollection := getOrderCollection()
	_, err = orderCollection.InsertOne(context.Background(), order)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create order"})
		return
	}

	cartCollection := getCartCollection()
	var cart Models.Cart
	err = cartCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&cart)
	if err == mongo.ErrNoDocuments {
		c.JSON(404, gin.H{"error": "Cart not found"})
		return
	}

	for _, selectedItem := range selectedItems.Items {
		for i, cartItem := range cart.Items {
			if selectedItem.ProductID == cartItem.ProductID {
				cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
				break
			}
		}
	}

	if len(cart.Items) == 0 {
		_, err = cartCollection.DeleteOne(context.Background(), bson.M{"user_id": userID})
	} else {
		cart.UpdatedAt = time.Now()
		_, err = cartCollection.UpdateOne(context.Background(), bson.M{"user_id": userID}, bson.M{"$set": cart})
	}

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update cart after order"})
		return
	}

	_, err = selectedItemsCollection.DeleteOne(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to clear selected items"})
		return
	}

	c.JSON(200, order)
}

func GetOrders(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	orderCollection := getOrderCollection()
	var orders []Models.Order
	cursor, err := orderCollection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get orders"})
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &orders); err != nil {
		c.JSON(500, gin.H{"error": "Failed to decode orders"})
		return
	}

	c.JSON(200, orders)
}

func GetAllOrders(c *gin.Context) {
	orderCollection := getOrderCollection()
	userCollection := getUserCollection()
	var orders []Models.Order

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := orderCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode orders"})
		return
	}

	for i, order := range orders {
		var user Models.User
		err := userCollection.FindOne(ctx, bson.M{"_id": order.UserID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user data"})
			return
		}
		orders[i].User = user
	}

	c.JSON(http.StatusOK, orders)
}

func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	var requestBody struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	orderCollection := getOrderCollection()
	var order Models.Order
	err = orderCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&order)
	if err == mongo.ErrNoDocuments {
		c.JSON(404, gin.H{"error": "Order not found"})
		return
	}

	if order.Status == "completed" || order.Status == "cancelled" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update a completed or cancelled order"})
		return
	}

	update := bson.M{"$set": bson.M{"status": requestBody.Status, "updatedAt": time.Now()}}
	_, err = orderCollection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully"})
}

func CancelOrder(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	orderID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid order ID format"})
		return
	}

	orderCollection := getOrderCollection()
	var order Models.Order
	err = orderCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&order)
	if err != nil {
		c.JSON(404, gin.H{"error": "Order not found"})
		return
	}

	if claims.Role == Middleware.Customer && order.UserID != claims.ID {
		c.JSON(403, gin.H{"error": "You are not authorized to cancel this order"})
		return
	}

	update := bson.M{"$set": bson.M{"status": "cancelled", "updatedAt": time.Now()}}
	_, err = orderCollection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to cancel order"})
		return
	}

	c.JSON(200, gin.H{"message": "Order cancelled successfully"})
}
