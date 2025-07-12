package sticks

type Hand struct {
	Fingers int // 0–4 (5+ becomes 0)
}

type Player struct {
	ID    string
	Hands [2]Hand
}

type Game struct {
	ID        string
	Players   [2]Player
	TurnIndex int    // 0 or 1
	Status    string // "waiting", "playing", "finished"
}
