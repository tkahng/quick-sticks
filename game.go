package sticks

import "errors"

type Game struct {
	ID        string
	Players   [2]*Player
	TurnIndex int    // 0 or 1
	Status    string // "waiting", "playing", "finished"
}

func (g *Game) PlayerCount() int {
	count := 0
	for _, player := range g.Players {
		if player != nil {
			count++
		}
	}
	return count
}

func NewGame(id string) *Game {
	return &Game{
		ID:        id,
		Players:   [2]*Player{},
		TurnIndex: 0,
		Status:    "waiting",
	}
}

func (g *Game) AddPlayer(p *Player) error {
	for idx, player := range g.Players {
		if player == nil {
			p.Idx = idx
			g.Players[idx] = p
			if idx == 1 {
				g.Status = "playing"
			} else {
				g.Status = "waiting"
			}
			return nil
		}
	}
	return errors.New("game is full")
}
