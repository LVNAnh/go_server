package Routes

import (
	"Server/Controllers"
	"Server/Middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// User routes
		api.POST("/register", Controllers.RegisterUser)
		api.POST("/login", Controllers.LoginUser)
		api.GET("/users", Middleware.AuthMiddleware(Middleware.Admin), Controllers.GetAllUsers)
		api.GET("/user/:id", Middleware.AuthMiddleware(Middleware.Admin), Controllers.GetUserByID)
		api.PUT("/user/:id", Middleware.AuthMiddleware(Middleware.Admin), Controllers.UpdateUser)
		api.DELETE("/user/:id", Middleware.AuthMiddleware(Middleware.Admin), Controllers.DeleteUser)

		// ProductCategory routes
		api.GET("/productcategories", Controllers.GetAllProductCategories)
		api.GET("/productcategory/:id", Controllers.GetProductCategoryByID)
		api.POST("/productcategory", Middleware.AuthMiddleware(Middleware.Admin), Controllers.CreateProductCategory)
		api.PUT("/productcategory/:id", Middleware.AuthMiddleware(Middleware.Admin), Controllers.UpdateProductCategory)
		api.DELETE("/productcategory/:id", Middleware.AuthMiddleware(Middleware.Admin), Controllers.DeleteProductCategory)

		// Product routes
		api.GET("/products", Controllers.GetAllProducts)
		api.GET("/product/:id", Controllers.GetProductByID)
		api.POST("/product", Middleware.AuthMiddleware(Middleware.Staff), Controllers.CreateProduct)
		api.PUT("/product/:id", Middleware.AuthMiddleware(Middleware.Staff), Controllers.UpdateProduct)
		api.DELETE("/product/:id", Middleware.AuthMiddleware(Middleware.Staff), Controllers.DeleteProduct)

		// ServiceCategory routes
		api.GET("/servicecategories", Controllers.GetAllServiceCategories)
		api.GET("/servicecategory/:id", Controllers.GetServiceCategoryByID)
		api.POST("/servicecategory", Middleware.AuthMiddleware(Middleware.Admin), Controllers.CreateServiceCategory)
		api.PUT("/servicecategory/:id", Middleware.AuthMiddleware(Middleware.Admin), Controllers.UpdateServiceCategory)
		api.DELETE("/servicecategory/:id", Middleware.AuthMiddleware(Middleware.Admin), Controllers.DeleteServiceCategory)

		// Service routes
		api.GET("/services", Controllers.GetAllServices)
		api.GET("/service/:id", Controllers.GetServiceByID)
		api.POST("/service", Middleware.AuthMiddleware(Middleware.Staff), Controllers.CreateService)
		api.PUT("/service/:id", Middleware.AuthMiddleware(Middleware.Staff), Controllers.UpdateService)
		api.DELETE("/service/:id", Middleware.AuthMiddleware(Middleware.Staff), Controllers.DeleteService)

		// Cart routes
		api.GET("/cart", Middleware.AuthMiddleware(Middleware.Customer), Controllers.GetCart)
		api.POST("/cart/add", Middleware.AuthMiddleware(Middleware.Customer), Controllers.AddToCart)
		api.DELETE("/cart/remove", Middleware.AuthMiddleware(Middleware.Customer), Controllers.RemoveFromCart)
		api.POST("/cart/update", Middleware.AuthMiddleware(Middleware.Customer), Controllers.UpdateCart)

		// Order routes
		api.POST("/order", Middleware.AuthMiddleware(Middleware.Customer), Controllers.CreateOrder)
		api.GET("/orders", Middleware.AuthMiddleware(Middleware.Customer), Controllers.GetOrders)
		api.DELETE("/order/:id", Middleware.AuthMiddleware(Middleware.Customer), Controllers.CancelOrder)

		// SelectedItems routes
		api.GET("/selecteditems", Middleware.AuthMiddleware(Middleware.Customer), Controllers.GetSelectedItems)
		api.POST("/selecteditems/add", Middleware.AuthMiddleware(Middleware.Customer), Controllers.AddToSelectedItems)
		api.POST("/selecteditems/addMultiple", Middleware.AuthMiddleware(Middleware.Customer), Controllers.AddMultipleToSelectedItems)
		api.DELETE("/selecteditems/remove", Middleware.AuthMiddleware(Middleware.Customer), Controllers.RemoveFromSelectedItems)
		api.POST("/selecteditems/update", Middleware.AuthMiddleware(Middleware.Customer), Controllers.UpdateSelectedItems)
		api.DELETE("/selecteditems/clear", Middleware.AuthMiddleware(Middleware.Customer), Controllers.ClearSelectedItems)

		// OrderBookingService routes
		api.POST("/orderbookingservice", Middleware.AuthMiddleware(Middleware.Customer), Controllers.CreateOrderBookingService)
		api.GET("/orderbookingservices", Middleware.AuthMiddleware(Middleware.Customer), Controllers.GetOrderBookingServices)
		api.PATCH("/orderbookingservice/:id/status", Middleware.AuthMiddleware(Middleware.Admin), Controllers.UpdateOrderBookingServiceStatus)
	}
}
