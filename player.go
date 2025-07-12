package sticks

type Player struct {
	ID    string
	Idx   int // 0 or 1
	Hands [2]*Hand
}

func (p *Player) Lives() int {
	lives := 0
	for _, hand := range p.Hands {
		if hand == nil {
			continue
		}
		lives += hand.Fingers
	}
	return lives
}
