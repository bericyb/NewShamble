package game

type Player struct {
	UserID   string
	Username string
}

type Game struct {
	ID      string
	Players []Player
	Mode    string
	Matches []Match
}

type Match struct {
	ID      string
	Player1 Player
	Player2 Player
	Winner  Player
}

func NewGame(id string, mode string) *Game {
	return &Game{
		ID:      id,
		Players: []Player{},
		Mode:    mode,
		Matches: []Match{
			{
				Player1: Player{},
				Player2: Player{},
				Winner:  Player{},
			},
			{
				Player1: Player{},
				Player2: Player{},
				Winner:  Player{},
			},
			{
				Player1: Player{},
				Player2: Player{},
				Winner:  Player{},
			},
		},
	}
}
