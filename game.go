package sticks

import (
	"fmt"
	"sync"
	"time"
)

// GameState represents the current state of the game
type GameState string

const (
	GameStateWaiting    GameState = "waiting"
	GameStateReady      GameState = "ready"
	GameStateInProgress GameState = "in_progress"
	GameStateFinished   GameState = "finished"
)

type GameInterface interface {
	AddPlayer(player *Player) error
	Attack(attackWithLeft bool, attackLeft bool) error
	GetCurrentPlayer() *Player
	GetOpponent() *Player
	Split(fromLeft bool, points int) error
	StartGame() error
	EndTurn()
	PrintScore()
}

type Game struct {
	ID          string
	Player1     *Player
	Player2     *Player
	CurrentTurn int       `json:"currentTurn"` // 0 for player1, 1 for player2
	State       GameState `json:"state"`
	Winner      *Player   `json:"winner,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	mutex       *sync.RWMutex
}

// PrintScore implements GameInterface.
func (g *Game) PrintScore() {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	fmt.Printf("Player 1: %d, %d\n", g.Player1.LeftHand.fingers, g.Player1.RightHand.fingers)
	fmt.Printf("Player 2: %d, %d\n", g.Player2.LeftHand.fingers, g.Player2.RightHand.fingers)
}

// GetTurn implements GameInterface.
func (g *Game) GetTurn() int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.CurrentTurn
}

var _ GameInterface = (*Game)(nil)

func NewGame(id string) *Game {
	return &Game{
		ID:          id,
		State:       GameStateWaiting,
		CreatedAt:   time.Now(),
		Player1:     nil,
		Player2:     nil,
		CurrentTurn: 0,
		Winner:      nil,
		mutex:       &sync.RWMutex{},
	}
}

func (g *Game) EndTurn() {
	g.CurrentTurn = 1 - g.CurrentTurn
}

// Attack implements GameInterface.
// Attack performs an attack move
func (g *Game) Attack(attackerIsLeft bool, defenderIsLeft bool) error {
	fmt.Printf("attacking\n")
	g.mutex.Lock()
	defer g.mutex.Unlock()

	fmt.Printf("locked\n")

	if g.State != GameStateInProgress {
		fmt.Printf("game is not in progress\n")
		return fmt.Errorf("game is not in progress")
	}

	fmt.Printf("getting players\n")

	// Get players directly without calling methods that acquire locks
	var attacker, defender *Player
	if g.CurrentTurn == 0 {
		attacker = g.Player1
		defender = g.Player2
	} else {
		attacker = g.Player2
		defender = g.Player1
	}

	fmt.Printf("got players\n")

	attackerHand := attacker.GetHand(attackerIsLeft)
	defenderHand := defender.GetHand(defenderIsLeft)

	// Perform attack
	err := attackerHand.Attack(defenderHand)
	if err != nil {
		return err
	}

	// Check if game is over
	if !defender.Alive() {
		g.State = GameStateFinished
		g.Winner = attacker
	} else {
		// Switch turns
		g.EndTurn()
	}

	return nil
}

// Split performs a split move
func (g *Game) Split(fromLeft bool, newLeftPoints int) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.State != GameStateInProgress {
		return fmt.Errorf("game is not in progress")
	}

	// Get current player directly without calling methods that acquire locks
	var player *Player
	if g.CurrentTurn == 0 {
		player = g.Player1
	} else {
		player = g.Player2
	}

	from := player.GetHand(fromLeft)
	other := player.GetHand(!fromLeft)
	err := other.Take(from, newLeftPoints)
	if err != nil {
		return err
	}

	// Switch turns
	g.EndTurn()

	return nil
}

// StartGame implements GameInterface.
func (g *Game) StartGame() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.State != GameStateReady {
		return fmt.Errorf("game is not ready to start")
	}

	if g.Player1 == nil || g.Player2 == nil {
		return fmt.Errorf("need two players to start")
	}

	g.State = GameStateInProgress
	g.CurrentTurn = 0 // Player1 starts
	return nil
}

func (g *Game) GetCurrentPlayer() *Player {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	if g.CurrentTurn == 0 {
		return g.Player1
	}
	return g.Player2
}

// GetOpponent returns the opponent of the current player
func (g *Game) GetOpponent() *Player {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	if g.CurrentTurn == 0 {
		return g.Player2
	}
	return g.Player1
}

func (g *Game) AddPlayer(player *Player) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.State != GameStateWaiting {
		return fmt.Errorf("game is not accepting players")
	}

	if g.Player1 == nil {
		g.Player1 = player
		return nil
	}

	if g.Player2 == nil {
		g.Player2 = player
		g.State = GameStateReady
		return nil
	}

	return fmt.Errorf("game is full")
}
