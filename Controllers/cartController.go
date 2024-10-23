package Controllers

import (
	"context"
	"time"

	"Server/Middleware"
	"Server/Models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func getCartCollection() *mongo.Collection {
	return Database.Collection("carts")
}

func getProductCollection() *mongo.Collection {
	return Database.Collection("products")
}

func AddToCart(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	var cartItem Models.CartItem
	if err := c.BindJSON(&cartItem); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	cartCollection := getCartCollection()
	var cart Models.Cart
	err := cartCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&cart)

	if err == mongo.ErrNoDocuments {
		cart = Models.Cart{
			UserID:    userID,
			Items:     []Models.CartItem{cartItem},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err = cartCollection.InsertOne(context.Background(), cart)
	} else {
		exists := false
		for i, item := range cart.Items {
			if item.ProductID == cartItem.ProductID {
				cart.Items[i].Quantity += cartItem.Quantity
				exists = true
				break
			}
		}
		if !exists {
			cart.Items = append(cart.Items, cartItem)
		}
		cart.UpdatedAt = time.Now()
		_, err = cartCollection.UpdateOne(context.Background(), bson.M{"user_id": userID}, bson.M{"$set": cart})
	}

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to add to cart"})
		return
	}

	c.JSON(200, cart)
}

func GetCart(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	cartCollection := getCartCollection()
	var cart Models.Cart
	err := cartCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&cart)

	if err != nil {
		c.JSON(404, gin.H{"error": "Cart not found"})
		return
	}

	productCollection := getProductCollection()
	for i, item := range cart.Items {
		var product Models.Product
		err := productCollection.FindOne(context.Background(), bson.M{"_id": item.ProductID}).Decode(&product)
		if err != nil {
			c.JSON(404, gin.H{"error": "Product not found"})
			return
		}

		cart.Items[i].Name = product.Name
		cart.Items[i].Price = product.Price
		cart.Items[i].ImageURL = product.ImageURL
	}

	c.JSON(200, cart)
}

func UpdateCart(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	var cartItem Models.CartItem
	if err := c.BindJSON(&cartItem); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	cartCollection := getCartCollection()
	var cart Models.Cart
	err := cartCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&cart)
	if err != nil {
		c.JSON(404, gin.H{"error": "Cart not found"})
		return
	}

	for i, item := range cart.Items {
		if item.ProductID == cartItem.ProductID {
			cart.Items[i].Quantity = cartItem.Quantity
			break
		}
	}

	cart.UpdatedAt = time.Now()
	_, err = cartCollection.UpdateOne(context.Background(), bson.M{"user_id": userID}, bson.M{"$set": cart})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update cart"})
		return
	}

	c.JSON(200, cart)
}

func RemoveFromCart(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	var cartItem Models.CartItem
	if err := c.BindJSON(&cartItem); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	cartCollection := getCartCollection()
	var cart Models.Cart
	err := cartCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&cart)
	if err == mongo.ErrNoDocuments {
		c.JSON(404, gin.H{"error": "Cart not found"})
		return
	}

	productRemoved := false
	for i, item := range cart.Items {
		if item.ProductID == cartItem.ProductID {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			productRemoved = true
			break
		}
	}

	if !productRemoved {
		c.JSON(400, gin.H{"error": "Product not found in cart"})
		return
	}

	if len(cart.Items) == 0 {
		_, err = cartCollection.DeleteOne(context.Background(), bson.M{"user_id": userID})
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete cart"})
			return
		}
	} else {
		cart.UpdatedAt = time.Now()
		_, err = cartCollection.UpdateOne(context.Background(), bson.M{"user_id": userID}, bson.M{"$set": cart})
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to update cart"})
			return
		}
	}

	c.JSON(200, gin.H{
		"message": "Product removed from cart",
		"cart":    cart,
	})
}
