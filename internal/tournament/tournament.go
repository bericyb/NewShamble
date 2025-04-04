package tournament

import (
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
)

type Tournament struct {
	ID             int
	MatchLobbies   []MatchLobby
	WaitingRoom    map[string]*Player
	Games          map[string]Game
	CurMatch       int
	WinnerUsername string
	CommandChan    chan GameCommand
	StartDate      time.Time
	Started        bool
}

type MatchLobby struct {
	Level    int
	Segments map[int][]*Player
}

type Game struct {
	Rounds         []Round
	Player1        *Player
	Player2        *Player
	WinnerUsername string
}

type Round struct {
	Player1Move string
	Player2Move string
	Winner      int
}

type Player struct {
	Username string
	MsgChan  chan GameResponse
	WinCount int
}

type GameCommand struct {
	Username string
	Command  string
	Payload  any
	Response chan GameResponse
}

type GameResponse struct {
	Command string
	Payload any
}

func NewTournament(id int) *Tournament {
	cmdChan := make(chan GameCommand, 100)
	t := &Tournament{
		ID:             id,
		MatchLobbies:   []MatchLobby{},
		WaitingRoom:    map[string]*Player{},
		Games:          map[string]Game{},
		CurMatch:       0,
		WinnerUsername: "",
		CommandChan:    cmdChan,
	}

	go t.Listen()
	return t
}

// Listen for game commands
func (t *Tournament) Listen() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Check if the StartDate has passed every minute
	go func(ticker *time.Ticker) {
		for {
			select {
			case <-ticker.C:
				if time.Now().After(t.StartDate) && !t.Started {
					slog.Info("Tournament start date has passed, starting tournament")

					t.CommandChan <- GameCommand{
						Command:  "startMatch",
						Response: nil,
					}
				}
			}
		}

	}(ticker)

	// Listen for commands
	for {
		select {
		case cmd := <-t.CommandChan:
			switch cmd.Command {
			case "move":
				t.AcceptPlayerMove(cmd.Username, cmd.Command)
			case "join":
				t.JoinWaitingRoom(cmd.Username, cmd.Payload.(*Player))
			case "leave":
				t.LeaveWaitingRoom(cmd.Username)
			case "startMatch":
				t.Start()
			case "endMatch":
				t.EndMatch()
			default:
				slog.Error("Unknown command", "command", cmd.Command)
			}
		}
	}
}

// Contains the logic to start the tournament, and also the go routine to check the every 20 seconds
func (t *Tournament) Start() {
	t.Started = true
	// Calculate num of matches in swiss format and create matches lobbies
	numMatches := math.Ceil(math.Log2(float64(len(t.WaitingRoom))))
	for i := range int(numMatches) + 1 {
		t.MatchLobbies = append(t.MatchLobbies, MatchLobby{
			Level:    i,
			Segments: map[int][]*Player{},
		})
	}

	// Add players to the first match
	// The first match only has 1 segment of 0 wins
	t.MatchLobbies[0].Segments[0] = []*Player{}
	for _, player := range t.WaitingRoom {
		t.MatchLobbies[0].Segments[0] = append(t.MatchLobbies[0].Segments[0], player)
	}

	// Create games for first Match
	t.CurMatch = 0

	// I guess the last player will be a bye in odd pairings
	// since the player will be nil gotta be careful not to deref them
	gameID := uuid.New().String()
	for i := range t.MatchLobbies[0].Segments[0] {
		if i%2 == 0 {
			t.Games[gameID] = Game{
				Rounds:  make([]Round, 5),
				Player1: t.MatchLobbies[0].Segments[0][i],
				Player2: nil,
			}
		} else {
			game, ok := t.Games[gameID]
			if !ok {
				slog.Error("Game not found during matchmaking! Huge problem!", "gameID", gameID)
				gameID = uuid.New().String()
				continue
			}
			game.Player2 = t.MatchLobbies[0].Segments[0][i]
			t.Games[gameID] = game
			gameID = uuid.New().String()
		}
	}

	// Alert players in games
	for _, game := range t.Games {
		if game.Player1 != nil {
			game.Player1.MsgChan <- GameResponse{
				Command: "gameStarted",
				Payload: game,
			}
		}
		if game.Player2 != nil {
			game.Player2.MsgChan <- GameResponse{
				Command: "gameStarted",
				Payload: game,
			}
		}
	}
	slog.Info("Tournament started", "numMatches", numMatches, "numPlayers", len(t.WaitingRoom))
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				slog.Info("Checking for game results")
				t.EndMatch()
			}
		}
	}()
}

// Start check every 20 seconds wrap up all games.
func (t *Tournament) EndMatch() {
	slog.Info("Locking tournament to check games")

	// Wrap up all games
	// increment win count for each winner
	for _, game := range t.Games {
		username := game.CalculateWinner()
		if username == "" {
			continue
		}
		// Update winner's win count
		wp := t.WaitingRoom[username]
		wp.WinCount++
		t.WaitingRoom[username] = wp
	}

	t.CurMatch++
	if t.CurMatch >= len(t.MatchLobbies) {
		// Tournament is over determine winner
		maxWins := 0
		winner := ""
		for _, player := range t.WaitingRoom {
			if player.WinCount > maxWins {
				maxWins = player.WinCount
				winner = player.Username
			}
		}
		t.WinnerUsername = winner
		slog.Info("Tournament ended", "winner", winner)
		// Alert players of tournament end
		for _, player := range t.WaitingRoom {
			player.MsgChan <- GameResponse{
				Command: "tournamentEnded",
				Payload: winner,
			}
		}
		return
	} else {
		// Calculate new segments for next match
		match := t.MatchLobbies[t.CurMatch]
		match.Segments = make(map[int][]*Player)
		for _, player := range t.WaitingRoom {
			match.Segments[player.WinCount] = append(match.Segments[player.WinCount], player)
		}
		t.MatchLobbies[t.CurMatch] = match
		// Create new games for current match
		for k, players := range t.MatchLobbies[t.CurMatch].Segments {
			gameID := uuid.New().String()
			for i := range players {
				if i%2 == 0 {
					t.Games[gameID] = Game{
						Rounds:  make([]Round, 5),
						Player1: t.MatchLobbies[t.CurMatch].Segments[k][i],
						Player2: nil,
					}
				} else {
					game, ok := t.Games[gameID]
					if !ok {
						slog.Error("Game not found during matchmaking in round! Huge problem!", "gameID", gameID, "round", t.CurMatch)
						gameID = uuid.New().String()
						continue
					}
					game.Player2 = t.MatchLobbies[t.CurMatch].Segments[k][i]
					t.Games[gameID] = game
					gameID = uuid.New().String()
				}
			}
		}
	}

	slog.Info("Finished generating games and alerting players")
	// Alert players in games
	for gID, game := range t.Games {
		if game.Player1 != nil {
			game.Player1.MsgChan <- GameResponse{
				Command: "gameStarted",
				Payload: gID,
			}
		}
		if game.Player2 != nil {
			game.Player2.MsgChan <- GameResponse{
				Command: "gameStarted",
				Payload: gID,
			}
		}
	}
}

func (t *Tournament) AcceptPlayerMove(username, move string) {
	gameID := uuid.New().String()
	// Check if game exists
	if game, ok := t.Games[gameID]; ok {
		//
		if game.Player1.Username == username {
			currentRound := 0
			for i := range game.Rounds {
				if game.Rounds[i].Player1Move == "" {
					game.Rounds[i].Player1Move = move
					currentRound = i

					// If other player has already played, check winner
					if game.Rounds[i].Player2Move != "" {
						game.Rounds[i].Winner = getWinner(game.Rounds[i].Player1Move, game.Rounds[i].Player2Move)
						p1wins, p2wins, _ := calculateStandings(game.Rounds)
						if p1wins >= 3 {
							game.WinnerUsername = game.Player1.Username
							game.Player1.MsgChan <- GameResponse{
								Command: "gameWon",
								Payload: game.Rounds[currentRound],
							}
							game.Player2.MsgChan <- GameResponse{
								Command: "gameLost",
								Payload: game.Rounds[currentRound],
							}
						} else if p2wins >= 3 {
							game.WinnerUsername = game.Player2.Username
							game.Player1.MsgChan <- GameResponse{
								Command: "gameLost",
								Payload: game.Rounds[currentRound],
							}
							game.Player2.MsgChan <- GameResponse{
								Command: "gameWon",
								Payload: game.Rounds[currentRound],
							}
						} else if i == len(game.Rounds)-1 {
							if p1wins == p2wins {
								game.WinnerUsername = ""
								game.Player1.MsgChan <- GameResponse{
									Command: "gameDraw",
									Payload: game.Rounds[currentRound],
								}
								game.Player2.MsgChan <- GameResponse{
									Command: "gameDraw",
									Payload: game.Rounds[currentRound],
								}
							} else if p1wins > p2wins {
								game.WinnerUsername = game.Player1.Username
								game.Player1.MsgChan <- GameResponse{
									Command: "gameWon",
									Payload: game.Rounds[currentRound],
								}
								game.Player2.MsgChan <- GameResponse{
									Command: "gameLost",
									Payload: game.Rounds[currentRound],
								}
							} else {
								game.WinnerUsername = game.Player2.Username
								game.Player1.MsgChan <- GameResponse{
									Command: "gameLost",
									Payload: game.Rounds[currentRound],
								}
								game.Player2.MsgChan <- GameResponse{
									Command: "gameWon",
									Payload: game.Rounds[currentRound],
								}
							}
						}
					} else {
						game.Player1.MsgChan <- GameResponse{
							Command: "moveAccepted",
							Payload: game.Rounds[currentRound],
						}
					}
					break
				}
			}
			// Rounds are up, there is no winner
			slog.Error("Player sent moves but rounds are up", "username", username)
		} else if game.Player2.Username == username {
			currentRound := 0
			for i := range game.Rounds {
				if game.Rounds[i].Player2Move == "" {
					game.Rounds[i].Player2Move = move
					currentRound = i

					// If other player has already played, check winner
					if game.Rounds[i].Player1Move != "" {
						game.Rounds[i].Winner = getWinner(game.Rounds[i].Player1Move, game.Rounds[i].Player2Move)
						p1wins, p2wins, _ := calculateStandings(game.Rounds)
						if p1wins >= 3 {
							game.WinnerUsername = game.Player1.Username
							game.Player1.MsgChan <- GameResponse{
								Command: "gameWon",
								Payload: game.Rounds[currentRound],
							}
							game.Player2.MsgChan <- GameResponse{
								Command: "gameLost",
								Payload: game.Rounds[currentRound],
							}
						}
						if p2wins >= 3 {
							game.WinnerUsername = game.Player2.Username
							game.Player1.MsgChan <- GameResponse{
								Command: "gameLost",
								Payload: game.Rounds[currentRound],
							}
							game.Player2.MsgChan <- GameResponse{
								Command: "gameWon",
								Payload: game.Rounds[currentRound],
							}
						} else if i == len(game.Rounds)-1 {
							if p1wins == p2wins {
								game.WinnerUsername = ""
								game.Player1.MsgChan <- GameResponse{
									Command: "gameDraw",
									Payload: game.Rounds[currentRound],
								}
								game.Player2.MsgChan <- GameResponse{
									Command: "gameDraw",
									Payload: game.Rounds[currentRound],
								}
							} else if p1wins > p2wins {
								game.WinnerUsername = game.Player1.Username
								game.Player1.MsgChan <- GameResponse{
									Command: "gameWon",
									Payload: game.Rounds[currentRound],
								}
								game.Player2.MsgChan <- GameResponse{
									Command: "gameLost",
									Payload: game.Rounds[currentRound],
								}
							} else {
								game.WinnerUsername = game.Player2.Username
								game.Player1.MsgChan <- GameResponse{
									Command: "gameLost",
									Payload: game.Rounds[currentRound],
								}
								game.Player2.MsgChan <- GameResponse{
									Command: "gameWon",
									Payload: game.Rounds[currentRound],
								}
							}
						}
					} else {
						game.Player2.MsgChan <- GameResponse{
							Command: "moveAccepted",
							Payload: game.Rounds[currentRound],
						}
					}
					break
				}
			}
			// Rounds are up, there is no winner
			slog.Error("Player sent moves but rounds are up", "username", username)
		} else {
			slog.Error("Player not found in game", "username", username)
			return
		}
		t.Games[gameID] = game
	}
}

func (t *Tournament) JoinWaitingRoom(username string, player *Player) {
	// Check if player already exists in waiting room
	if _, ok := t.WaitingRoom[username]; !ok {
		t.WaitingRoom[username] = player
	}
}

func (t *Tournament) LeaveWaitingRoom(username string) {
	// Check if player exists in waiting room
	if _, ok := t.WaitingRoom[username]; ok {
		delete(t.WaitingRoom, username)
	}
}

// After 5 games if there are all draws, we will not have a winner
func (g *Game) CalculateWinner() string {
	// Check if we already have a winner
	if g.WinnerUsername != "" {
		return g.WinnerUsername
	}
	if (g.Player1 == nil) && (g.Player2 == nil) {
		return ""
	}
	if g.Player1 == nil {
		return g.Player2.Username
	}
	if g.Player2 == nil {
		return g.Player1.Username
	}

	// If not, count the rounds
	p1 := 0
	p2 := 0
	for _, round := range g.Rounds {
		winner := getWinner(round.Player1Move, round.Player2Move)
		if winner == 1 {
			p1++
		}
		if winner == 2 {
			p2++
		}
	}
	if p1 > p2 {
		g.WinnerUsername = g.Player1.Username
	}
	if p2 > p1 {
		g.WinnerUsername = g.Player2.Username
	}
	if p1 == p2 {
		g.WinnerUsername = ""
	}
	return g.WinnerUsername
}

func getWinner(move1, move2 string) int {
	// No one played
	if move1 == "" && move2 != "" {
		return 0
	}

	if move1 != "" {
		return 2
	}

	if move2 != "" {
		return 1
	}

	if move1 == move2 {
		return 0
	}

	if (move1 == "rock" && move2 == "scissors") || (move1 == "scissors" && move2 == "paper") || (move1 == "paper" && move2 == "rock") {
		return 1
	}

	return 2
}

// Get the standings of a game's rounds
func calculateStandings(rounds []Round) (int, int, int) {
	p1wins := 0
	p2wins := 0
	draws := 0
	for _, round := range rounds {
		if round.Winner == 1 {
			p1wins++
		} else if round.Winner == 2 {
			p2wins++
		} else {
			draws++
		}
	}
	return p1wins, p2wins, draws

}
