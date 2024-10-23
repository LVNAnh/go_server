package Controllers

import (
	"context"
	"net/http"

	"Server/Middleware"
	"Server/Models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getCollection(name string) *mongo.Collection {
	return Database.Collection(name)
}

func CreateProductCategory(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	if claims.Role > Middleware.Staff {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to create a product category"})
		return
	}

	var productCategory Models.ProductCategory
	if err := c.ShouldBindJSON(&productCategory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	productCategory.ID = primitive.NewObjectID()

	collection := getCollection("product_categories")
	_, err := collection.InsertOne(context.Background(), productCategory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, productCategory)
}

func GetAllProductCategories(c *gin.Context) {
	var productCategories []Models.ProductCategory
	collection := getCollection("product_categories")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var productCategory Models.ProductCategory
		cursor.Decode(&productCategory)
		productCategories = append(productCategories, productCategory)
	}

	c.JSON(http.StatusOK, productCategories)
}

func GetProductCategoryByID(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var productCategory Models.ProductCategory
	collection := getCollection("product_categories")
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&productCategory)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product category not found"})
		return
	}

	c.JSON(http.StatusOK, productCategory)
}

func UpdateProductCategory(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	if claims.Role > Middleware.Staff {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to update a product category"})
		return
	}

	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var productCategory Models.ProductCategory
	if err := c.ShouldBindJSON(&productCategory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	collection := getCollection("product_categories")
	update := bson.M{"$set": bson.M{
		"name":        productCategory.Name,
		"description": productCategory.Description,
	}}
	result, err := collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product category not found"})
		return
	}

	c.JSON(http.StatusOK, productCategory)
}

func DeleteProductCategory(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	if claims.Role != Middleware.Admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete a product category"})
		return
	}

	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	collection := getCollection("product_categories")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product category not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
