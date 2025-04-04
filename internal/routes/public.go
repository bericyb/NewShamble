package routes

import (
	"Roshamble/internal/handlers"

	"github.com/gin-gonic/gin"
)

func PublicRoutes(r *gin.Engine, handler *handlers.Handler) {
	// Public routes
	r.GET("/ping", handler.PingHandler)
	r.GET("/empty", handler.Empty)

	r.GET("/auth/logout", handler.Logout)

	// Get login page
	r.GET("/auth/login", handler.GetLogin)

	// Submit OTP to verify user and set cookie
	r.POST("/auth/login", handler.Login)

	// Post phone number to trigger otp send
	r.POST("/auth/otp", handler.TriggerOTP)

	r.GET("/redirect", handler.Redirect)
}
