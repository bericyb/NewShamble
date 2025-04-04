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

	// Root
	auth.GET("/", handler.GetDashboard)

	// Game handlers
	auth.GET("/play/:tournamentID", handler.GetPlay)
	auth.GET("/ws/play/:tournamentID", handler.WsHandler)
	auth.POST("/leave/:tournamentID", handler.LeaveTournament)

	// Profile handlers
	auth.GET("/profile", handler.GetProfile)
	auth.PATCH("/profile", handler.UpdateProfile)

	// Hall of fame handlers
	auth.GET("/halloffame", handler.GetPastTournaments)

}
