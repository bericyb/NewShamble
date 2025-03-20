package handlers

import (
	"Roshamble/internal/game"
	"Roshamble/internal/services"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func getClaims(c *gin.Context) (services.Claims, error) {
	claimsValue, ok := c.Get("claims")
	if !ok {
		fmt.Println("error getting claims")
		return services.Claims{}, fmt.Errorf("error getting claims")
	}
	claims, ok := claimsValue.(jwt.MapClaims)
	if !ok {
		fmt.Println("error converting claims")
		return services.Claims{}, fmt.Errorf("error getting claims")
	}

	structClaims := services.Claims{
		ID:       claims["id"].(string),
		Username: claims["username"].(string),
		Email:    claims["email"].(string),
	}
	return structClaims, nil
}

func (h *Handler) GetDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{})
}

func (h *Handler) GetPlay(c *gin.Context) {
	currentPlayers := 0

	row := h.DB.QueryRow("SELECT (SELECT COUNT(*) FROM quickplay_queue) + (SELECT COUNT(*) FROM ranked_queue) + (SELECT COUNT(*) FROM tournament_queue)")

	err := row.Scan(&currentPlayers)
	if err != nil {
		slog.Error("Error querying for total online count", "error", err)
	}

	c.HTML(http.StatusOK, "play.html", gin.H{"CurrentPlayers": currentPlayers})
}

func (h *Handler) DeclineMatch(c *gin.Context) {

	userData, err := getClaims(c)
	if err == nil {
		_, err := h.DB.Exec("DELETE FROM quickplay_queue WHERE user_id = $1", userData.ID)
		if err != nil {
			slog.Error("Error updating removing player from queue", "error", err)
		}
	}

	c.Header("HX-Location", "/play")
	c.HTML(http.StatusOK, "redirector", gin.H{"Title": "Matchmaking Failed", "Message": "You have declined the match."})
}

func (h *Handler) GetQueue(c *gin.Context) {
	// Capitalize the mode for display purposes
	mode := c.Param("mode")
	if mode == "" {
		c.Redirect(http.StatusFound, "/play")
		return
	}

	formattedMode := cases.Title(language.English).String(mode)

	userData, err := getClaims(c)
	if err != nil {
		c.Redirect(http.StatusUnauthorized, "index.html")
		return
	}

	currentPlayers := 1
	// insert the player into the respective queue and return the queue size for current players
	if mode == "quickplay" {
		// Insert the user into the quickplay queue with their ID
		_, err = h.DB.Exec("INSERT INTO quickplay_queue (user_id) VALUES ($1) ON CONFLICT (user_id) DO UPDATE SET queued_at = NOW(), match_id = NULL, matched_at = NULL", userData.ID)
		if err != nil {
			slog.Error("Error inserting into quickplay queue")
			slog.Error(err.Error())
			c.HTML(http.StatusInternalServerError, "play.html", gin.H{"error": "Database error"})
			return
		}
		row := h.DB.QueryRow("SELECT COUNT(*) FROM quickplay_queue")
		err = row.Scan(&currentPlayers)
		if err != nil {
			slog.Error("Error querying quickplay queue")
			slog.Error(err.Error())
			c.HTML(http.StatusInternalServerError, "play.html", gin.H{"error": "Database error"})
			return
		}
	} else if mode == "ranked" {
		_, err = h.DB.Exec("INSERT INTO ranked_queue (user_id) VALUES ($1) ON CONFLICT (user_id) DO UPDATE SET queued_at = NOW(), match_id = NULL, matched_at = NULL", userData.ID)
		if err != nil {
			slog.Error("Error inserting into ranked queue")
			slog.Error(err.Error())
			c.HTML(http.StatusInternalServerError, "play.html", gin.H{"error": "Database error"})
			return
		}
		row := h.DB.QueryRow("SELECT COUNT(*) FROM ranked_queue")
		err = row.Scan(&currentPlayers)
		if err != nil {
			slog.Error("Error querying ranked queue")
			slog.Error(err.Error())
			c.HTML(http.StatusInternalServerError, "play.html", gin.H{"error": "Database error"})
			return
		}
	} else if mode == "tournament" {
		_, err = h.DB.Exec("INSERT INTO tournament_queue (user_id) VALUES ($1) ON CONFLICT (user_id) DO UPDATE SET queued_at = NOW(), match_id = NULL, matched_at = NULL", userData.ID)
		if err != nil {
			slog.Error("Error inserting into tournament queue")
			slog.Error(err.Error())
			c.HTML(http.StatusInternalServerError, "play.html", gin.H{"error": "Database error"})
			return
		}
		row := h.DB.QueryRow("SELECT COUNT(*) FROM tournament_queue")
		err = row.Scan(&currentPlayers)
		if err != nil {
			slog.Error("Error querying tournament queue")
			slog.Error(err.Error())
			c.HTML(http.StatusInternalServerError, "play.html", gin.H{"error": "Database error"})
			return
		}
	} else {
		slog.Error("Invalid game mode")
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid game mode"})
		return
	}

	c.HTML(http.StatusOK, "queue.html", gin.H{"Mode": c.Param("mode"), "formattedMode": formattedMode, "CurrentPlayers": currentPlayers, "PlayerID": userData.ID})
}

func (h *Handler) GetQueueStatus(c *gin.Context) {
	mode := c.Param("mode")
	if mode == "" {
		c.Redirect(http.StatusFound, "/play")
		return
	}

	var currentStatus string
	var matchID sql.NullString
	if mode == "quickplay" {
		// Check if the player has a game ready
		row := h.DB.QueryRow("SELECT match_id FROM quickplay_queue WHERE user_id = $1", c.Param("playerID"))
		err := row.Scan(&matchID)

		if err != nil {
			slog.Error("Error querying quickplay queue")
			slog.Error(err.Error())
			c.HTML(http.StatusInternalServerError, "play.html", gin.H{"error": "Matchmaking error. Please leave and try again..."})
			return
		}
		// If the player has a game, render the match_ready page
		if matchID.Valid {
			rows, err := h.DB.Query("SELECT users.id, users.username FROM quickplay_queue JOIN users on users.id = quickplay_queue.user_id WHERE quickplay_queue.match_id = $1", matchID.String)
			if err != nil {
				slog.Error("Error querying for match users' usernames", "match_id", matchID.String)
				slog.Error(err.Error())
				c.HTML(http.StatusInternalServerError, "play.html", gin.H{"error": "Matchmaking error. Please leave and try again..."})
				return
			}

			players := []game.Player{}
			for rows.Next() {
				var username string
				var user_id string
				if err := rows.Scan(&user_id, &username); err != nil {
					slog.Error("Error scanning username or id")
					slog.Error(err.Error())
					c.HTML(http.StatusInternalServerError, "play.html", gin.H{"error": "Matchmaking error. Please leave and try again..."})
					return
				}
				players = append(players, game.Player{UserID: user_id, Username: username})
			}

			c.HTML(http.StatusOK, "match_ready.html", gin.H{"MatchReady": true, "Player1": players[0], "Player2": players[1], "MatchID": matchID.String})
			return
		}
	} else if mode == "ranked" {
		// Check if the player has a game ready
		row := h.DB.QueryRow("SELECT status, match_id FROM ranked_queue WHERE user_id = $1", c.Param("playerID"))
		err := row.Scan(&currentStatus, &matchID)
		if err != nil {
			slog.Error("Error querying ranked queue")
			slog.Error(err.Error())
			c.HTML(http.StatusInternalServerError, "play.html", gin.H{"error": "Matchmaking error. Please leave and try again..."})
			return
		}
		// If the player has a game, render the match_ready page
		if currentStatus != "pending" {
			c.HTML(http.StatusOK, "match_ready.html", gin.H{"MatchReady": true, "PlayerID": c.Param("playerID"), "MatchID": matchID})
			return
		}
	} else if mode == "tournament" {
		// Check if the player has a game ready
		row := h.DB.QueryRow("SELECT status, match_id FROM tournament_queue WHERE user_id = $1", c.Param("playerID"))
		err := row.Scan(&currentStatus, &matchID)
		if err != nil {
			slog.Error("Error querying tournament queue")
			slog.Error(err.Error())
			c.HTML(http.StatusInternalServerError, "play.html", gin.H{"error": "Matchmaking error. Please leave and try again..."})
			return
		}
		// If the player has a game, render the match_ready page
		if currentStatus != "pending" {
			c.HTML(http.StatusOK, "match_ready.html", gin.H{"MatchReady": true, "PlayerID": c.Param("playerID"), "MatchID": matchID})
			return
		}
	} else {
		slog.Error("Invalid game mode")
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid game mode"})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{})
}

func (h *Handler) DeletePlayerFromQueue(c *gin.Context) {
	userData, err := getClaims(c) // Ensure claims are retrieved to validate user
	if err != nil {
		c.Redirect(http.StatusUnauthorized, "index.html")
		return
	}

	if userData.ID != c.Param("playerID") {
		c.Redirect(http.StatusUnauthorized, "index.html")
		return
	}

	mode := c.Param("mode")
	if mode == "quickplay" {
		_, err = h.DB.Exec("DELETE FROM quickplay_queue WHERE user_id = $1", userData.ID)
	} else if mode == "ranked" {
		_, err = h.DB.Exec("DELETE FROM ranked_queue WHERE user_id = $1", userData.ID)
	} else if mode == "tournament" {
		_, err = h.DB.Exec("DELETE FROM tournament_queue WHERE user_id = $1", userData.ID)
	} else {
		slog.Error("Invalid game mode")
		c.HTML(http.StatusBadRequest, "redirector", gin.H{"error": "Invalid game mode", "URL": "/play"})
		return
	}

	c.Header("HX-Location", "/play")
	c.HTML(http.StatusOK, "play.html", gin.H{"message": "You have been removed from the queue."})
}

func (h *Handler) GetMatch(c *gin.Context) {
	matchID := c.Param("gameID")
	if matchID == "" {
		c.HTML(http.StatusBadRequest, "redirector", gin.H{"Message": "Invalid game ID", "URL": "/play"})
		return
	}

	// row := h.DB.QueryRow("SELECT * FROM matches WHERE id = $1", matchID)
	// var match game.Match
	// err := row.Scan(&match.ID, &match.Player1, &match.Player2, &match.Winner)
	// if err != nil {
	// 	slog.Error("Error querying match")
	// 	slog.Error(err.Error())
	// 	c.HTML(http.StatusInternalServerError, "redirector", gin.H{"Message": "Match not found", "URL": "/play"})
	// 	return
	// }

	c.HTML(http.StatusOK, "game.html", gin.H{"Match": matchID})
}
