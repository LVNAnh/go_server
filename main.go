package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"Server/Controllers"
	"Server/Routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

	Routes.SetupRoutes(router)

	router.StaticFS("/uploads", http.Dir("./uploads"))

	port := "8080"
	fmt.Printf("Server running at http://localhost:%s\n", port)
	log.Fatal(router.Run(":" + port))
}
