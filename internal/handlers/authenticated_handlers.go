package handlers

import (
	"Roshamble/internal/services"
	"Roshamble/internal/tournament"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
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
	}
	return structClaims, nil
}

func (h *Handler) GetDashboard(c *gin.Context) {
	claims, err := getClaims(c)
	if err != nil {
		slog.Error("Error getting claims", "error", err)
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	tournamentData := services.TournamentData{}

	tournamentData, err = services.GetTournamentData(c, h.DB, claims)
	if err != nil {
		slog.Error("Error getting tournament data", "error", err)
	}

	slog.Info("open tournament", "openID", tournamentData.OpenTournament.ID)
	slog.Info("ongoing tournament", "ongoing", tournamentData.OngoingTournament.ID)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{"Claims": claims, "OpenTournament": tournamentData.OpenTournament, "OpenTournamentCountDown": "30 seconds", "OngoingTournament": tournamentData.OngoingTournament, "LastTournamentDate": tournamentData.LastTournament.GetDateTimeString(), "UpcomingTournaments": tournamentData.UpcomingTournaments})
}

func (h *Handler) GetPastTournaments(c *gin.Context) {
	var pastTournaments []services.Tournament

	pastTournaments, err := services.GetPastTournaments(c, h.DB)
	if err != nil {
		slog.Error("Error getting past tournament data", "error", err)
	}
	slog.Info("Tournament data ", "past tourney", pastTournaments)

	c.HTML(http.StatusOK, "halloffame.html", gin.H{"Tournaments": pastTournaments})
}

func (h *Handler) GetProfile(c *gin.Context) {
	claims, err := getClaims(c)
	if err != nil || claims.ID == "" {
		slog.Error("Error getting claims", "error", err)
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	p, err := services.GetUserProfile(c, h.DB, claims.ID)
	if err != nil {
		slog.Error("Error getting user notification settings", "error", err)
		c.Redirect(http.StatusFound, "/")
		return
	}

	c.HTML(http.StatusOK, "profile.html", gin.H{"Username": p.Username, "Email": p.Email, "NewTournaments": p.NewTournamentsNotif, "FriendsJoined": p.FriendsJoinedNotif, "TournamentStarting": p.TournamentStartingNotif})
	return
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	claims, err := getClaims(c)
	if err != nil {
		slog.Error("Error getting claims", "error", err)
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	message, errorMessage := services.UpdateUserProfile(c, h.DB, claims)
	c.HTML(http.StatusOK, "profile.html", gin.H{"Message": message, "ErrorMessage": errorMessage})
	return
}

func (h *Handler) GetPlay(c *gin.Context) {
	_, err := getClaims(c)
	if err != nil {
		slog.Error("Error getting claims", "error", err)
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	dbT, err := services.GetTournament(c, h.DB)
	if err != nil {
		slog.Error("Error getting tournament data", "error", err)
	}

	if dbT.ID == 0 {
		slog.Error("Error getting tournament data", "error", "tournament not found")
		return
	}

	// Initialize tournament if it doesn't exist
	st, ok := h.Tournaments.Load(dbT.ID)
	if !ok {
		slog.Info("Tournament not found in memory, creating new tournament")
		st = tournament.NewTournament(dbT.ID)
		h.Tournaments.Store(dbT.ID, st)
		st, ok = h.Tournaments.Load(dbT.ID)
		if !ok {
			slog.Error("Error storing new tournament in memory")
			return
		}
	}

	c.HTML(http.StatusOK, "play.html", gin.H{"Error": "", "Countdown": "", "InviteLink": "", "Tournament": dbT})
}

func (h *Handler) LeaveTournament(c *gin.Context) {
	claims, err := getClaims(c)
	if err != nil {
		slog.Error("Error getting claims", "error", err)
		c.Redirect(http.StatusFound, "/auth/login")
	}

	tIDStr := c.Param("tournamentID")
	tID, err := strconv.Atoi(tIDStr)
	if err != nil {
		slog.Error("Error parsing tournamentID from url param", "error", err.Error())
	}

	err = services.RemoveTournamentPlayer(c, h.DB, tID, claims)
	if err != nil {
		slog.Error("Error removing player from tournament", "error", err.Error())
	}
	c.HTML(http.StatusOK, "redirector.html", gin.H{"Title": "Leaving tournament", "Message": "You may join again before the tournament starts"})
}

var upgrader = websocket.Upgrader{
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		slog.Error("TODO: Fix ws error hanlding", "status", status, "reason", reason)
	},
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Add origin check
		return true
	},
	EnableCompression: false,
}

func (h *Handler) WsHandler(c *gin.Context) {

	claims, err := getClaims(c)
	if err != nil {
		slog.Error("Error getting claims", "error", err)
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	tournamentID := c.Param("tournamentID")
	if tournamentID == "" {
		slog.Error("Error getting tournament ID from url param")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	defer conn.Close()
	if err != nil {
		slog.Error("Error upgrading websocket connection", "error", err)
		return
	}
	slog.Info("Websocket connection established")

	tID, err := strconv.Atoi(tournamentID)
	if err != nil {
		slog.Error("Error parsing tournamentID from url param", "error", err.Error())
		return
	}
	st, ok := h.Tournaments.Load(tID)
	if !ok {
		slog.Error("Tournament not found...")
		return
	}

	recvChan := make(chan tournament.GameResponse)
	player := tournament.Player{
		Username: claims.Username,
		MsgChan:  recvChan,
		WinCount: 0,
	}

	t := st.(*tournament.Tournament)
	t.CommandChan <- tournament.GameCommand{
		Username: claims.Username,
		Command:  "join",
		Payload:  &player,
	}
	slog.Info("Player joined tournament", "username", player.Username)

	var data map[string]any

	go func() {
		for {
			msg := []byte{}
			select {
			case cmd := <-recvChan:
				switch cmd.Command {
				case "gameWon":
					msg = []byte(fmt.Sprintf("You won the game! %s", cmd.Payload))

				case "gameLost":
					msg = []byte(fmt.Sprintf("You lost the game! %s", cmd.Payload))

				case "gameDraw":
					msg = []byte(fmt.Sprintf("The game ended in a draw! %s", cmd.Payload))

				case "moveAccepted":
					msg = []byte(fmt.Sprintf("Your move was accepted! %s Waiting for other player...", cmd.Payload))

				case "gameStarted":
					msg = []byte(fmt.Sprintf("The game has started! GameID: %s", cmd.Payload))

				case "gameEnded":
					msg = []byte(fmt.Sprintf("The game has ended! %s", cmd.Payload))
				case "tournamentEnded":
					msg = []byte(fmt.Sprintf("The tournament has ended! %s is the winner!", cmd.Payload))

				default:
					slog.Error("Unknown command received from tournament", "command", cmd.Command)
				}
			}
			err = conn.WriteMessage(websocket.BinaryMessage, msg)
			if err != nil {
				slog.Error("Error writing message", "error", err)
				return
			}
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			slog.Error("Error reading message", "error", err)
			return
		}
		slog.Info("Message received", "message", string(msg))

		err = json.Unmarshal(msg, &data)
		if err != nil {
			slog.Error("Error unmarshalling message", "error", err)
			return
		}

		if move, ok := data["move"].(string); ok {
			slog.Info("Move received", "move", move)
			t.CommandChan <- tournament.GameCommand{
				Username: claims.Username,
				Command:  "move",
				Payload:  move,
			}

		}

	}
}
