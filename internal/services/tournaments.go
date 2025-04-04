package services

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

type TournamentData struct {
	OpenTournament      Tournament
	OngoingTournament   Tournament
	LastTournament      Tournament
	UpcomingTournaments []Tournament
}

type Tournament struct {
	ID             int    `form:"id"`
	Name           string `form:"name"`
	Description    string `form:"description"`
	Emoji          string `form:"emoji"`
	Prize          string `form:"prize"`
	PrizeURL       string `form:"prize_url"`
	InviteLevel    int    `form:"invite_level"`
	StartDate      string `form:"start_date"`
	Location       string `form:"location"`
	WinnerID       string
	WinnerUsername string
}

func (t *Tournament) GetDateTimeString() string {
	res, err := time.Parse("2006-01-02T15:04:05.000000Z", t.StartDate)
	if err != nil {
		slog.Error("Error parsing tournament start date", "error", err.Error())
		return t.StartDate
	}
	return res.Format("January 2 : 3:04 PM")
}

func GetTournamentData(c *gin.Context, db *sql.DB, claims Claims) (TournamentData, error) {
	tournamentData := TournamentData{}
	tournamentQueue := []Tournament{}

	rows, err := db.Query("SELECT id, name, COALESCE(description, ''), prize, COALESCE(prize_url, ''), COALESCE(emoji, ''), start_date FROM tournaments WHERE winner_id is NULL AND invite_level <= $1 AND start_date > NOW() ORDER BY start_date ASC LIMIT 4", claims.InviteLevel)
	if err != nil {
		return tournamentData, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Tournament
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Prize, &t.Emoji, &t.PrizeURL, &t.StartDate); err != nil {
			return tournamentData, err
		}
		tournamentQueue = append(tournamentQueue, t)
	}

	if err := rows.Err(); err != nil {
		return tournamentData, err
	}

	lastTournament := Tournament{}
	row := db.QueryRow("SELECT tournaments.id, name, winner_id, users.username, start_date, prize, COALESCE(prize_url, ''), COALESCE(emoji, '') FROM tournaments JOIN users on users.id = winner_id WHERE winner_id IS NOT NULL ORDER BY start_date DESC LIMIT 1")

	if err := row.Scan(&lastTournament.ID, &lastTournament.Name, &lastTournament.WinnerID, &lastTournament.WinnerUsername, &lastTournament.StartDate, &lastTournament.Prize, &lastTournament.PrizeURL, &lastTournament.Emoji); err != nil && err != sql.ErrNoRows {
		slog.Error("Error scanning last tournament", slog.Any("error", err))
	}

	tournamentData.LastTournament = lastTournament

	if len(tournamentQueue) > 0 {
		slog.Info("Got tournaments queued up")
		tournamentData.OpenTournament = tournamentQueue[0]
		slog.Info("Set openTournament to ", "open", tournamentData.OpenTournament)
		if len(tournamentQueue) > 1 {
			tournamentData.UpcomingTournaments = tournamentQueue[1:]
		}
	} else {
		tournamentData.OpenTournament = Tournament{}
	}

	tournamentData.OngoingTournament = Tournament{}
	row = db.QueryRow("SELECT id, name, start_date, prize, COALESCE(prize_url, ''), COALESCE(emoji,'') FROM tournaments WHERE winner_id IS NULL AND start_date >= NOW() ORDER BY start_date DESC LIMIT 1")
	if err := row.Scan(&tournamentData.OngoingTournament.ID, &tournamentData.OngoingTournament.Name, &tournamentData.OngoingTournament.Prize, &tournamentData.OngoingTournament.PrizeURL, &tournamentData.OngoingTournament.Emoji, &tournamentData.OngoingTournament.StartDate); err != nil && err != sql.ErrNoRows {
		slog.Error("Error scanning ongoing tournament", slog.Any("error", err))
	}

	return tournamentData, nil
}

func GetTournament(c *gin.Context, db *sql.DB) (Tournament, error) {
	t := Tournament{}

	tID := c.Param("tournamentID")

	row := db.QueryRow("SELECT id, name, prize, COALESCE(prize_url, ''), start_date FROM tournaments WHERE id = $1", tID)

	err := row.Scan(&t.ID, &t.Name, &t.Prize, &t.PrizeURL, &t.StartDate)
	if err != nil {
		slog.Error("Error scanning tournament by id", "error", err.Error())
	}

	return t, err
}

func CreateTournament(c *gin.Context, db *sql.DB) error {
	tournament := Tournament{}

	if err := c.Bind(&tournament); err != nil {
		slog.Error("Error binding tournament data", slog.Any("error", err))
		return err
	}
	_, err := db.Exec("INSERT INTO tournaments (name, start_date) VALUES ($1, $2, $3)",
		tournament.Prize, tournament.StartDate, tournament.PrizeURL, tournament.Emoji)

	if err != nil {
		slog.Error("Error inserting tournament into database", slog.Any("error", err))
	}

	return err
}

func GetPastTournaments(c *gin.Context, db *sql.DB) ([]Tournament, error) {
	var pt []Tournament
	rows, err := db.Query("SELECT tournaments.id, name, COALESCE(description, ''), COALESCE(emoji, ''), prize, COALESCE(prize_url, ''), start_date, winner_id, users.username  FROM tournaments JOIN users ON users.id = tournaments.winner_id WHERE winner_id IS NOT NULL")
	if err != nil {
		slog.Error("Error fetching past tournaments", "error", err.Error())
		return pt, err
	}

	for rows.Next() {
		tr := Tournament{}
		err := rows.Scan(&tr.ID, &tr.Name, &tr.Description, &tr.Emoji, &tr.Prize, &tr.PrizeURL, &tr.StartDate, &tr.WinnerID, &tr.WinnerUsername)
		if err != nil {
			slog.Error("Error scanning past tournament row")
			return pt, err
		}
		pt = append(pt, tr)
	}

	return pt, err
}

func RemoveTournamentPlayer(c *gin.Context, db *sql.DB, tournamentID int, claims Claims) error {
	_, err := db.Exec("DELETE FROM tournament_players WHERE tournament_id = $1 AND player_id = $2", tournamentID, claims.ID)
	if err != nil {
		slog.Error("Error removing player from tournament", "error", err.Error())
	}

	return err
}

func AddTournamentPlayer(c *gin.Context, db *sql.DB, tournamentID int, claims Claims) error {
	_, err := db.Exec("INSERT INTO tournament_players (tournament_id, player_id) VALUES tournament_id = $1, player_id = $2", tournamentID, claims.ID)
	if err != nil {
		slog.Error("Error adding player to tournament", "error", err.Error())
	}

	return err
}
