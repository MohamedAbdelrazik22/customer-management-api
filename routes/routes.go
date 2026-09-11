package routes

import (
	"github.com/gin-gonic/gin"

	"customer-management-api/handlers"
	"customer-management-api/middleware"
)

// SetupRoutes registers all API routes on the given Gin engine.
func SetupRoutes(r *gin.Engine, customerHandler *handlers.CustomerHandler, authHandler *handlers.AuthHandler) {
	// Public route — no token required
	r.POST("/login", authHandler.Login)

	// Public read-only routes
	r.GET("/customers", customerHandler.GetAll)
	r.GET("/customers/:id", customerHandler.GetByID)

	// Protected routes — JWT token required
	protected := r.Group("/customers")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("", customerHandler.Create)
		protected.PUT("/:id", customerHandler.Update)
		protected.DELETE("/:id", customerHandler.Delete)
	}
}
