package services

import (
	"database/sql"
	"log/slog"
	"math/rand"
	"os"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type RegisterRequest struct {
	Phone    string `form:"phone"`
	Username string `form:"username,omitempty"`
	Email    string `form:"email,omitempty"`
}

type Claims struct {
	ID          string
	Username    string
	Phone       string
	InviteLevel int
}

type OTPRequest struct {
	Phone string `form:"phone"`
}

func TriggerOTP(c *gin.Context, db *sql.DB) (string, string, string) {
	req := OTPRequest{}

	err := c.Bind(&req)
	if err != nil {
		return "", "", "Bad request. Please try again later"
	}

	// To limit the number of requests, check if the phone number already has a verification code created in the last 2 minutes
	row := db.QueryRow("SELECT created_at FROM otp WHERE phone = $1 AND created_at > NOW() - INTERVAL '2 minutes'", req.Phone)

	if err := row.Scan(&time.Time{}); err == nil {
		return req.Phone, "A verification code has already been sent to this phone numberuser.sdfj Please wait a moment before requesting another.", ""
	} else if err != sql.ErrNoRows {
		slog.Error("Error checking for existing OTP", slog.Any("error", err))
		return "", "", "There was a problem. Please try again later"
	}

	// Generate a random 6-digit verification code
	code := 100000 + rand.Intn(900000)
	// Store the code in the database
	_, err = db.Exec("INSERT INTO otp (phone, code) VALUES ($1, $2) ON CONFLICT (phone) DO UPDATE SET code = $2, created_at = NOW()", req.Phone, code)
	if err != nil {
		slog.Error("Failed to insert or update OTP", slog.Any("error", err))
		return "", "", "There was a problem sending the verification code. Please try again later"
	}

	// TODO: Send the verification code to the user's phone number using an SMS service
	// For now, we'll just log it

	slog.Info("Verification code sent", slog.String("phone", req.Phone), slog.Int("code", code))

	return req.Phone, "Check your messages for a verification code", ""
}

type VerificationRequest struct {
	Phone string `form:"phone"`
	Code  int    `form:"otp"`
}

func VerifyUser(c *gin.Context, db *sql.DB) (string, string) {
	req := VerificationRequest{}

	if err := c.Bind(&req); err != nil {
		slog.Error("Error binding request", slog.Any("error", err))
		return "", "Bad request. Please try again later"
	}

	slog.Info("User phone", slog.String("phone", req.Phone), slog.Int("code", req.Code))

	// Retrieve the stored verification code for the given phone number and make sure it's not expired
	// Assuming the code is valid for 5 minutes
	row := db.QueryRow("SELECT code FROM otp WHERE phone = $1 AND created_at > NOW() - INTERVAL '5 minutes'", req.Phone)
	var storedCode int
	if err := row.Scan(&storedCode); err != nil {
		if err == sql.ErrNoRows {
			return "", "No verification code found for this phone number"
		}
		slog.Error("Error scanning stored code", slog.Any("error", err))
		return "", "There was a problem verifying your phone number. Please try again later"
	}
	if storedCode != req.Code {
		return "", "Invalid verification code"
	}

	// If the user exists with this phone number, retrieve their details
	// Otherwise, create a new user with a random username
	row = db.QueryRow("SELECT id, username, invite_level FROM users WHERE phone = $1", req.Phone)
	claims := Claims{}
	if err := row.Scan(&claims.ID, &claims.Username, &claims.InviteLevel); err != nil {
		if err == sql.ErrNoRows {
			// Create a new user with a random username
			claims.Username = petname.Generate(2, "-")
			_, err := db.Exec("INSERT INTO users (phone, username) VALUES ($1, $2)", req.Phone, claims.Username)
			row := db.QueryRow("SELECT id FROM users WHERE phone = $1", req.Phone)
			if err := row.Scan(&claims.ID); err != nil {
				slog.Error("Error creating new user", slog.Any("error", err))
				return "", "There was a sigining into your account. Please log in again"
			}
			if err != nil {
				slog.Error("Error creating new user", slog.Any("error", err))
				return "", "There was a problem creating your account. Please try again later"
			}
		} else {
			slog.Error("Error scanning user details", slog.Any("error", err))
			return "", "There was a problem logging in to your account. Please try again later"
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":        claims.ID,
		"username":  claims.Username,
		"invitelvl": claims.InviteLevel, // Default invite level
		"exp":       time.Now().Add(time.Hour * 24 * 365).Unix(),
	})

	secretKey := []byte(os.Getenv("SECRET"))
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		slog.Error("Error signing token", slog.Any("error", err))
		return "", "There was a problem signing in. Please try again later"
	}

	return "Bearer " + tokenString, ""
}

type Profile struct {
	Username                string `form:"username,omitempty"`
	Email                   string `form:"email"`
	NewTournamentsNotif     bool   `form:"newTournaments"`
	FriendsJoinedNotif      bool   `form:"friendsJoined"`
	TournamentStartingNotif bool   `form:"tournamentStarting"`
}

func GetUserProfile(c *gin.Context, db *sql.DB, userID string) (Profile, error) {
	var p Profile
	row := db.QueryRow("SELECT username, COALESCE(email, '') AS email, new_tournaments_notif, friends_joined_notif, tournament_starting_notif FROM users WHERE id = $1", userID)
	if err := row.Scan(&p.Username, &p.Email, &p.NewTournamentsNotif, &p.FriendsJoinedNotif, &p.TournamentStartingNotif); err != nil {
		if err != sql.ErrNoRows {
			slog.Error("Error scanning notification settings", slog.Any("error", err))
		}
		return p, err
	}
	return p, nil
}

func UpdateUserProfile(c *gin.Context, db *sql.DB, claims Claims) (string, string) {
	req := Profile{}

	if err := c.Bind(&req); err != nil {
		slog.Error("Error binding request", slog.Any("error", err))
		return "", "Bad request. Please try again later"
	}

	// Update the user's profile
	_, err := db.Exec("UPDATE users SET username = COALESCE(NULLIF($1, ''), username), email = COALESCE(NULLIF($2, ''), email), new_tournaments_notif = $3, friends_joined_notif = $4, tournament_starting_notif = $5 WHERE id = $6", req.Username, req.Email, req.NewTournamentsNotif, req.TournamentStartingNotif, req.FriendsJoinedNotif, claims.ID)
	if err != nil {
		slog.Error("Error updating user profile", slog.Any("error", err))
		return "", "There was a problem updating your profile. Please try again later"
	}

	return "Profile updated", ""
}
