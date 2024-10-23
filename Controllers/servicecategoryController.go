package Controllers

import (
	"Server/Middleware"
	"Server/Models"
	"context"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getServiceCategoryCollection() *mongo.Collection {
	return Database.Collection("service_categories")
}

func CreateServiceCategory(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	if claims.Role != Middleware.Admin && claims.Role != Middleware.Staff {
		c.JSON(403, gin.H{"error": "Permission denied"})
		return
	}

	var serviceCategory Models.ServiceCategory
	if err := c.ShouldBindJSON(&serviceCategory); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	serviceCategory.ID = primitive.NewObjectID()
	collection := getServiceCategoryCollection()
	if _, err := collection.InsertOne(context.Background(), serviceCategory); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, serviceCategory)
}

func GetAllServiceCategories(c *gin.Context) {
	var serviceCategories []Models.ServiceCategory
	collection := getServiceCategoryCollection()
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var serviceCategory Models.ServiceCategory
		cursor.Decode(&serviceCategory)
		serviceCategories = append(serviceCategories, serviceCategory)
	}

	c.JSON(200, serviceCategories)
}

func GetServiceCategoryByID(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID format"})
		return
	}

	var serviceCategory Models.ServiceCategory
	collection := getServiceCategoryCollection()
	if err := collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&serviceCategory); err != nil {
		c.JSON(404, gin.H{"error": "Service category not found"})
		return
	}

	c.JSON(200, serviceCategory)
}

func UpdateServiceCategory(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	if claims.Role != Middleware.Admin && claims.Role != Middleware.Staff {
		c.JSON(403, gin.H{"error": "Permission denied"})
		return
	}

	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID format"})
		return
	}

	var serviceCategory Models.ServiceCategory
	if err := c.ShouldBindJSON(&serviceCategory); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	collection := getServiceCategoryCollection()
	update := bson.M{
		"$set": bson.M{
			"name":        serviceCategory.Name,
			"description": serviceCategory.Description,
		},
	}
	result, err := collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(404, gin.H{"error": "Service category not found"})
		return
	}

	c.JSON(200, serviceCategory)
}

func DeleteServiceCategory(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)
	if claims.Role != Middleware.Admin && claims.Role != Middleware.Staff {
		c.JSON(403, gin.H{"error": "Permission denied"})
		return
	}

	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID format"})
		return
	}

	collection := getServiceCategoryCollection()
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(404, gin.H{"error": "Service category not found"})
		return
	}

	c.Status(204)
}
