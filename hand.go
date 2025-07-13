package sticks

import (
	"errors"
)

type HandInterface interface {
	Alive() bool
	Attack(opp *Hand) error
	Set(num int)
	Take(other *Hand, points int) error
}
type Hand struct {
	fingers int // 0–4 (5+ becomes 0)
}

func (h *Hand) Set(num int) {
	h.fingers = num
}

func (h *Hand) Attack(opp *Hand) error {
	if opp == nil {
		return errors.New("opponent hand is nil")
	}
	if !opp.Alive() {
		return errors.New("opponent hand is dead")
	}
	opp.fingers += h.fingers
	return nil
}

func (h *Hand) Alive() bool {
	return h.fingers < 5
}

func (h *Hand) Take(other *Hand, points int) error {
	// check if this hand can take that many points without dying
	if h.fingers+points > 5 {
		return errors.New("this is more points than this hand can take")
	}
	if other.fingers-points < 0 {
		return errors.New("the other hand does not have enough points")
	}
	h.fingers += points
	other.fingers -= points
	return nil
}

func NewHand() *Hand {
	return &Hand{
		fingers: 1,
	}
}
