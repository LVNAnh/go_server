package main

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"Server/Controllers"
	"Server/Routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func uploadToCloudinary(file multipart.File, fileName string) (string, error) {
	cld, err := cloudinary.NewFromParams("dflhancsp", "437865386617669", "uLJSc-9ItdeXSbWyEndQ3x-F1FY")
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{PublicID: fileName})
	if err != nil {
		return "", err
	}
	return uploadResult.SecureURL, nil
}

func main() {
	clientOptions := options.Client().ApplyURI("mongodb+srv://NHATANH:WMGaAVaGCtsPnC1k@cluster0.6z5yhqo.mongodb.net/golangproject")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB: ", err)
	}

	database := client.Database("golang_project")
	Controllers.Database = database

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:6969"},
		AllowMethods:     []string{"POST", "GET", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image not found"})
			return
		}

		fileContent, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read file"})
			return
		}
		defer fileContent.Close()

		url, err := uploadToCloudinary(fileContent, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": url})
	})

	Routes.SetupRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server running at http://localhost:%s\n", port)
	log.Fatal(router.Run(":" + port))
}
