package routes

import (
	"Roshamble/internal/handlers"

	"github.com/gin-gonic/gin"
)

func PublicRoutes(r *gin.Engine, handler *handlers.Handler) {
	// Public routes
	r.GET("/", handler.Root)
	r.GET("/ping", handler.PingHandler)
	r.GET("/empty", handler.Empty)
	r.GET("/auth/login", handler.GetLogin)
	r.GET("/auth/logout", handler.Logout)
	r.POST("/auth/login", handler.Login)
	r.GET("/auth/register", handler.GetRegister)
	r.POST("/auth/register", handler.Register)
	r.GET("/redirect", handler.Redirect)
}
