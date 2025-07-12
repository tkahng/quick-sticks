package sticks

import (
	"errors"
)

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
