package Controllers

import (
	"context"
	"net/http"
	"time"

	"Server/Middleware"
	"Server/Models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func getSelectedItemsCollection() *mongo.Collection {
	return Database.Collection("selected_items")
}

func AddToSelectedItems(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	var selectedItem Models.SelectedItem
	if err := c.ShouldBindJSON(&selectedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	productCollection := getProductCollection()
	var product Models.Product
	if err := productCollection.FindOne(context.Background(), bson.M{"_id": selectedItem.ProductID}).Decode(&product); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	selectedItem.Name = product.Name
	selectedItem.ImageURL = product.ImageURL

	collection := getSelectedItemsCollection()
	var selectedItems Models.SelectedItems
	if err := collection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&selectedItems); err == mongo.ErrNoDocuments {
		selectedItems = Models.SelectedItems{
			UserID:    userID,
			Items:     []Models.SelectedItem{selectedItem},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if _, err := collection.InsertOne(context.Background(), selectedItems); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to selected items"})
			return
		}
	} else {
		exists := false
		for i, item := range selectedItems.Items {
			if item.ProductID == selectedItem.ProductID {
				selectedItems.Items[i].Quantity += selectedItem.Quantity
				exists = true
				break
			}
		}
		if !exists {
			selectedItems.Items = append(selectedItems.Items, selectedItem)
		}
		selectedItems.UpdatedAt = time.Now()
		if _, err := collection.UpdateOne(context.Background(), bson.M{"user_id": userID}, bson.M{"$set": selectedItems}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update selected items"})
			return
		}
	}
	c.JSON(http.StatusOK, selectedItems)
}

func GetSelectedItems(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	collection := getSelectedItemsCollection()
	var selectedItems Models.SelectedItems
	if err := collection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&selectedItems); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Selected items not found"})
		return
	}

	c.JSON(http.StatusOK, selectedItems)
}

func UpdateSelectedItems(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	var selectedItem Models.SelectedItem
	if err := c.ShouldBindJSON(&selectedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	collection := getSelectedItemsCollection()
	var selectedItems Models.SelectedItems
	if err := collection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&selectedItems); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Selected items not found"})
		return
	}

	for i, item := range selectedItems.Items {
		if item.ProductID == selectedItem.ProductID {
			selectedItems.Items[i].Quantity = selectedItem.Quantity
			break
		}
	}
	selectedItems.UpdatedAt = time.Now()

	if _, err := collection.UpdateOne(context.Background(), bson.M{"user_id": userID}, bson.M{"$set": selectedItems}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update selected items"})
		return
	}

	c.JSON(http.StatusOK, selectedItems)
}

func RemoveFromSelectedItems(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	var selectedItem Models.SelectedItem
	if err := c.ShouldBindJSON(&selectedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	collection := getSelectedItemsCollection()
	var selectedItems Models.SelectedItems
	if err := collection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&selectedItems); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Selected items not found"})
		return
	}

	for i, item := range selectedItems.Items {
		if item.ProductID == selectedItem.ProductID {
			selectedItems.Items = append(selectedItems.Items[:i], selectedItems.Items[i+1:]...)
			break
		}
	}

	if len(selectedItems.Items) == 0 {
		if _, err := collection.DeleteOne(context.Background(), bson.M{"user_id": userID}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete selected items"})
			return
		}
	} else {
		selectedItems.UpdatedAt = time.Now()
		if _, err := collection.UpdateOne(context.Background(), bson.M{"user_id": userID}, bson.M{"$set": selectedItems}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update selected items"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product removed from selected items",
		"items":   selectedItems,
	})
}

func ClearSelectedItems(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	collection := getSelectedItemsCollection()
	if _, err := collection.DeleteOne(context.Background(), bson.M{"user_id": userID}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear selected items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Selected items cleared"})
}

func AddMultipleToSelectedItems(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	userID := claims.ID

	var selectedItems []Models.SelectedItem
	if err := c.ShouldBindJSON(&selectedItems); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	collection := getSelectedItemsCollection()
	var existingSelectedItems Models.SelectedItems
	if err := collection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&existingSelectedItems); err == mongo.ErrNoDocuments {
		newSelectedItems := Models.SelectedItems{
			UserID:    userID,
			Items:     selectedItems,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if _, err := collection.InsertOne(context.Background(), newSelectedItems); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add selected items"})
			return
		}
	} else {
		updatedItems := existingSelectedItems.Items
		for _, newItem := range selectedItems {
			exists := false
			for i, item := range updatedItems {
				if item.ProductID == newItem.ProductID {
					updatedItems[i].Quantity += newItem.Quantity
					exists = true
					break
				}
			}
			if !exists {
				updatedItems = append(updatedItems, newItem)
			}
		}
		existingSelectedItems.Items = updatedItems
		existingSelectedItems.UpdatedAt = time.Now()
		if _, err := collection.UpdateOne(context.Background(), bson.M{"user_id": userID}, bson.M{"$set": existingSelectedItems}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update selected items"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Items added successfully",
		"items":   existingSelectedItems.Items,
	})
}
