package Controllers

import (
	"context"
	"net/http"
	"strconv"

	"Server/Middleware"
	"Server/Models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateService(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)

	if claims.Role != Middleware.Admin && claims.Role != Middleware.Staff {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to create services"})
		return
	}

	var service Models.Service
	err := c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse multipart form"})
		return
	}

	file, err := c.FormFile("image")
	if err == nil {
		fileContent, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open file"})
			return
		}
		defer fileContent.Close()

		// Upload to Cloudinary
		url, err := uploadToCloudinary(fileContent, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upload image to Cloudinary"})
			return
		}
		service.ImageURL = url
	}

	service.Name = c.PostForm("name")
	service.Price, _ = strconv.ParseFloat(c.PostForm("price"), 64)
	service.Description = c.PostForm("description")
	service.ServiceCategory, _ = primitive.ObjectIDFromHex(c.PostForm("servicecategory"))

	if service.Name == "" || service.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	service.ID = primitive.NewObjectID()

	collection := getCollection("services")
	_, err = collection.InsertOne(context.Background(), service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, service)
}

func GetAllServices(c *gin.Context) {
	var services []Models.Service
	collection := getCollection("services")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var service Models.Service
		cursor.Decode(&service)
		services = append(services, service)
	}

	c.JSON(http.StatusOK, services)
}

func GetServiceByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var service Models.Service
	collection := getCollection("services")
	err = collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&service)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.JSON(http.StatusOK, service)
}

func UpdateService(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)

	if claims.Role != Middleware.Admin && claims.Role != Middleware.Staff {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update services"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var existingService Models.Service
	collection := getCollection("services")

	err = collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&existingService)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	err = c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse multipart form"})
		return
	}

	file, err := c.FormFile("image")
	if err == nil {
		fileContent, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open file"})
			return
		}
		defer fileContent.Close()

		// Upload to Cloudinary
		url, err := uploadToCloudinary(fileContent, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upload image to Cloudinary"})
			return
		}
		existingService.ImageURL = url
	}

	if name := c.PostForm("name"); name != "" {
		existingService.Name = name
	}
	if price, err := strconv.ParseFloat(c.PostForm("price"), 64); err == nil && price > 0 {
		existingService.Price = price
	}
	if description := c.PostForm("description"); description != "" {
		existingService.Description = description
	}
	if category := c.PostForm("servicecategory"); category != "" {
		if serviceCategory, err := primitive.ObjectIDFromHex(category); err == nil {
			existingService.ServiceCategory = serviceCategory
		}
	}

	if existingService.Name == "" || existingService.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"name":            existingService.Name,
			"price":           existingService.Price,
			"description":     existingService.Description,
			"servicecategory": existingService.ServiceCategory,
			"imageurl":        existingService.ImageURL,
		},
	}

	result, err := collection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.JSON(http.StatusOK, existingService)
}

func DeleteService(c *gin.Context) {
	claims := c.MustGet("user").(*Middleware.UserClaims)

	if claims.Role != Middleware.Admin {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete services"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	collection := getCollection("services")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
