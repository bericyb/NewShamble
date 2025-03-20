package routes

import (
	"Roshamble/internal/auth"
	"Roshamble/internal/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProtectedRoutes(r *gin.Engine, handler *handlers.Handler) {
	// Protected routes
	auth := r.Group("/").Use(auth.JwtAuthMiddleware())

	auth.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "You are authenticated"})
	})

	auth.GET("/dashboard", handler.GetDashboard)
	auth.GET("/play", handler.GetPlay)
	auth.GET("/decline-match", handler.DeclineMatch)
	auth.GET("/queue/:mode", handler.GetQueue)
	auth.DELETE("/queue/:mode/:playerID", handler.DeletePlayerFromQueue)
	auth.POST("/queuestatus/:mode/:playerID", handler.GetQueueStatus)
	auth.GET("/match/:gameID", handler.GetMatch)
}
