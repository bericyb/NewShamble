package handlers

import (
	"Roshamble/internal/game"
	"Roshamble/internal/services"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	DB    *sql.DB
	Games map[string]game.Game
}

func (h *Handler) Root(c *gin.Context) {
	if cookie, err := c.Cookie("Authorization"); err != nil && cookie != "" {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{})
	}
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func (h *Handler) Empty(c *gin.Context) {
	c.HTML(http.StatusOK, "empty.html", gin.H{})
}

func (h *Handler) GetLogin(c *gin.Context) {
	if cookie, err := c.Cookie("Authorization"); err != nil && cookie != "" {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{})
	}
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func (h *Handler) Login(c *gin.Context) {
	token, errMessage := services.LoginUser(c, h.DB)
	if errMessage != "" {
		c.HTML(http.StatusOK, "login.html", gin.H{"message": errMessage})
	}

	c.SetCookie("Authorization", token, 3600*24*365, "/", "", false, true)
	c.Header("HX-Redirect", "/dashboard")
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) GetRegister(c *gin.Context) {
	if cookie, err := c.Cookie("Authorization"); err != nil && cookie != "" {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{})
	}
	c.HTML(http.StatusOK, "register.html", gin.H{})
}

func (h *Handler) Register(c *gin.Context) {
	token, err := services.RegisterUser(c, h.DB)
	if err != "" {
		c.HTML(http.StatusOK, "register.html", gin.H{"message": err})
	}

	c.SetCookie("Authorization", token, 3600*24*365, "/", "", false, true)
	c.Header("HX-Redirect", "/dashboard")
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("Authorization", "", -1, "/", "", false, true)
	c.Header("HX-Redirect", "/")
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func (h *Handler) PingHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "pong"})
}

func (h *Handler) Redirect(c *gin.Context) {
	message := c.Param("message")
	title := c.Param("title")
	c.HTML(http.StatusOK, "redirector.html", gin.H{"Message": message, "Title": title})
}
