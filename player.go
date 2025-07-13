package sticks

type Player struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	LeftHand  *Hand  `json:"leftHand"`
	RightHand *Hand  `json:"rightHand"`
}

func (p *Player) Alive() bool {
	return p.LeftHand.Alive() ||
		p.RightHand.Alive()
}

func (p *Player) GetHand(isLeft bool) *Hand {
	if isLeft {
		return p.LeftHand
	}
	return p.RightHand
}

func NewPlayer(id, name string) *Player {
	return &Player{
		ID:        id,
		Name:      name,
		LeftHand:  NewHand(),
		RightHand: NewHand(),
	}
}
