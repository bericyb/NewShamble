package services

import (
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string `form:"username"`
	Email    string `form:"email"`
	Password string `form:"password"`
}

type Claims struct {
	ID       string
	Username string
	Email    string
	password string
}

func RegisterUser(c *gin.Context, db *sql.DB) (string, string) {
	req := RegisterRequest{}
	if err := c.Bind(&req); err != nil {
		slog.Error(err.Error())
		return "", "Bad request, please try again later"
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error(err.Error())
		return "", "There was a problem creating your account. Please try again later"
	}

	// Insert user into database
	_, err = db.Exec("INSERT INTO users (email, username, password) VALUES ($1, $2, $3)",
		req.Email, req.Username, string(hashedPassword))

	if err != nil {
		slog.Error("Failed to create user")
		slog.Error(err.Error())
		return "", "Username or email is taken"
	}

	row := db.QueryRow("SELECT id, username, email, password FROM users WHERE (username = $1 or email = $1)", req.Username)

	claims := Claims{}

	if err := row.Scan(&claims.ID, &claims.Username, &claims.Email, &claims.password); err != nil {
		if err == sql.ErrNoRows {
			return "", "Invalid username, email, or password"
		} else {
			slog.Error("Error scanning claims")
			slog.Error(err.Error())
			return "", "There was a problem creating your account. Please try again later"
		}
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(claims.password), []byte(req.Password)); err != nil {
		return "", "Invalid username, email, or password"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       claims.ID,
		"username": claims.Username,
		"email":    claims.Email,
		"exp":      time.Now().Add(time.Hour * 24 * 365).Unix(),
	})

	secretKey := []byte(os.Getenv("SECRET"))
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		slog.Error("Error signing token", slog.Any("error", err))
		return "", "There was a problem signing in. Please try again later"
	}

	return "Bearer " + tokenString, ""
}

type LoginRequest struct {
	UsernameOrEmail string `form:"username_or_email"`
	Password        string `form:"password"`
}

func LoginUser(c *gin.Context, db *sql.DB) (string, string) {

	req := LoginRequest{}

	err := c.Bind(&req)
	if err != nil {
		return "", "Bad request. Please try again later"
	}

	row := db.QueryRow("SELECT id, username, email, password FROM users WHERE (username = $1 or email = $1)", req.UsernameOrEmail)

	claims := Claims{}

	if err := row.Scan(&claims.ID, &claims.Username, &claims.Email, &claims.password); err != nil {
		if err == sql.ErrNoRows {
			return "", "Invalid username, email, or password"
		} else {
			slog.Error("Error scanning claims")
			slog.Error(err.Error())
			return "", "There was a problem logging in to your account. Please try again later"
		}
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(claims.password), []byte(req.Password)); err != nil {
		return "", "Invalid username, email, or password"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       claims.ID,
		"username": claims.Username,
		"email":    claims.Email,
		"exp":      time.Now().Add(time.Hour * 24 * 365).Unix(),
	})

	secretKey := []byte(os.Getenv("SECRET"))
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		slog.Error("Error signing token", slog.Any("error", err))
		return "", "There was a problem signing in. Please try again later"
	}

	return "Bearer " + tokenString, ""
}
