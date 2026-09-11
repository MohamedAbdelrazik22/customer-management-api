package routes

import (
	"github.com/gin-gonic/gin"

	"customer-management-api/handlers"
)

// SetupRoutes registers all API routes on the given Gin engine.
func SetupRoutes(r *gin.Engine, customerHandler *handlers.CustomerHandler) {
	// Group all customer endpoints under /customers
	customers := r.Group("/customers")
	{
		customers.GET("", customerHandler.GetAll)
		customers.GET("/:id", customerHandler.GetByID)
		customers.POST("", customerHandler.Create)
		customers.PUT("/:id", customerHandler.Update)
		customers.DELETE("/:id", customerHandler.Delete)
	}
}
