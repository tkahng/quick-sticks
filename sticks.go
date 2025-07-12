package sticks

import "errors"

type Hand struct {
	Fingers int // 0–4 (5+ becomes 0)
}

func (h *Hand) Attack(opp *Hand) error {
	if opp.Fingers == 0 {
		return errors.New("opponent has no fingers")
	}
	h.Fingers++
	opp.Fingers--
	return nil
}

func (h *Hand) Split(other *Hand) error {
	if other.Fingers > 1 {
		half := other.Fingers / 2
		h.Fingers += half
		other.Fingers -= half
		return nil
	}
	return errors.New("cannot split. not enough other fingers")
}

type Player struct {
	ID    string
	Idx   int // 0 or 1
	Hands [2]*Hand
}

type Game struct {
	ID        string
	Players   [2]*Player
	TurnIndex int    // 0 or 1
	Status    string // "waiting", "playing", "finished"
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
