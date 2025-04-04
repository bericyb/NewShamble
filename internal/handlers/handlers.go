package handlers

import (
	"Roshamble/internal/services"
	"database/sql"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	DB          *sql.DB
	Tournaments sync.Map // map[int]*tournament.Tournament
}

func (h *Handler) Empty(c *gin.Context) {
	c.HTML(http.StatusOK, "empty.html", gin.H{})
}

func (h *Handler) GetLogin(c *gin.Context) {
	if cookie, err := c.Cookie("Authorization"); err != nil && cookie != "" {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{})
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{})
	return
}

func (h *Handler) TriggerOTP(c *gin.Context) {
	phone, message, errMessage := services.TriggerOTP(c, h.DB)
	if errMessage != "" {
		c.HTML(http.StatusOK, "login.html", gin.H{"message": errMessage})
		return
	}
	c.HTML(http.StatusOK, "otp.html", gin.H{"Message": message, "Phone": phone})
}

func (h *Handler) Login(c *gin.Context) {
	token, errMessage := services.VerifyUser(c, h.DB)
	slog.Info("User logged in", "token", token, "errMessage", errMessage)
	if errMessage != "" {
		c.HTML(http.StatusOK, "login.html", gin.H{"Error": errMessage})
		return
	}

	c.SetCookie("Authorization", token, 3600*24*365, "/", "", false, true)
	c.Header("HX-Redirect", "/")
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
