package sticks

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// MatchmakingRequest represents a player's request to join a game
type MatchmakingRequest struct {
	Player   *Player
	Response chan *MatchmakingResponse
}

// MatchmakingResponse contains the result of matchmaking
type MatchmakingResponse struct {
	Game  *Game
	Error error
}

// GameBroker handles matchmaking and game lifecycle management
type GameBroker struct {
	// Configuration
	maxConcurrentGames int
	matchmakingTimeout time.Duration
	gameTimeout        time.Duration

	// Matchmaking queue
	queue chan *MatchmakingRequest

	// Active games tracking
	activeGames map[string]*GameSession
	gamesMutex  *sync.RWMutex

	// Concurrency control
	gameSemaphore chan struct{} // Limits concurrent games

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

// GameSession wraps a game with its goroutine management
type GameSession struct {
	Game      *Game
	Context   context.Context
	Cancel    context.CancelFunc
	StartTime time.Time
	Players   []*Player
}

// NewGameBroker creates a new game broker
func NewGameBroker(maxConcurrentGames int) *GameBroker {
	ctx, cancel := context.WithCancel(context.Background())

	return &GameBroker{
		maxConcurrentGames: maxConcurrentGames,
		matchmakingTimeout: 30 * time.Second,
		gameTimeout:        30 * time.Minute,
		queue:              make(chan *MatchmakingRequest, 1000), // Buffered queue
		activeGames:        make(map[string]*GameSession),
		gameSemaphore:      make(chan struct{}, maxConcurrentGames),
		ctx:                ctx,
		cancel:             cancel,
		gamesMutex:         new(sync.RWMutex),
		wg:                 new(sync.WaitGroup),
	}
}

// Start begins the matchmaking broker
func (gb *GameBroker) Start() {
	gb.wg.Add(3)

	// Start matchmaking goroutine
	go gb.matchmakingWorker()

	// Start game cleanup goroutine
	go gb.gameCleanupWorker()

	// Start metrics/monitoring goroutine
	go gb.monitoringWorker()

	log.Printf("GameBroker started with max %d concurrent games", gb.maxConcurrentGames)
}

// Stop gracefully shuts down the broker
func (gb *GameBroker) Stop() {
	gb.cancel()
	close(gb.queue)
	gb.wg.Wait()

	// Clean up remaining games
	gb.gamesMutex.Lock()
	for _, session := range gb.activeGames {
		session.Cancel()
	}
	gb.gamesMutex.Unlock()

	log.Printf("GameBroker stopped")
}

// RequestGame adds a player to the matchmaking queue
func (gb *GameBroker) RequestGame(player *Player) (*Game, error) {
	responseChan := make(chan *MatchmakingResponse, 1)

	request := &MatchmakingRequest{
		Player:   player,
		Response: responseChan,
	}

	// Try to add to queue with timeout
	select {
	case gb.queue <- request:
		// Successfully queued
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("matchmaking queue is full")
	case <-gb.ctx.Done():
		return nil, fmt.Errorf("broker is shutting down")
	}

	// Wait for response
	select {
	case response := <-responseChan:
		return response.Game, response.Error
	case <-time.After(gb.matchmakingTimeout):
		return nil, fmt.Errorf("matchmaking timeout")
	case <-gb.ctx.Done():
		return nil, fmt.Errorf("broker is shutting down")
	}
}

// matchmakingWorker handles the core matchmaking logic
func (gb *GameBroker) matchmakingWorker() {
	defer gb.wg.Done()

	var waitingPlayer *MatchmakingRequest

	for {
		select {
		case request, ok := <-gb.queue:
			if request == nil {
				// Channel closed
				return
			}
			if !ok {
				// Channel closed
				return
			}

			if waitingPlayer == nil {
				// First player waiting
				waitingPlayer = request
				log.Printf("Player %s waiting for match", request.Player.ID)
			} else {
				// Second player arrived, create game
				gb.createGame(waitingPlayer, request)
				waitingPlayer = nil
			}

		case <-gb.ctx.Done():
			// Send cancellation to waiting player
			if waitingPlayer != nil {
				waitingPlayer.Response <- &MatchmakingResponse{
					Error: fmt.Errorf("matchmaking cancelled"),
					Game:  nil,
				}
			}
			return
		}
	}
}

// createGame creates a new game between two players
func (gb *GameBroker) createGame(player1Req, player2Req *MatchmakingRequest) {
	// Check if we can create a new game (concurrency limit)
	select {
	case gb.gameSemaphore <- struct{}{}:
		// Got slot, proceed
	default:
		// No slots available
		player1Req.Response <- &MatchmakingResponse{
			Error: fmt.Errorf("server at capacity"),
			Game:  nil,
		}
		player2Req.Response <- &MatchmakingResponse{
			Error: fmt.Errorf("server at capacity"),
			Game:  nil,
		}
		return
	}

	// Create game
	gameID := fmt.Sprintf("game_%d", time.Now().UnixNano())
	game := NewGame(gameID)

	// Add players to game
	if err := game.AddPlayer(player1Req.Player); err != nil {
		gb.respondWithError(player1Req, player2Req, err)
		<-gb.gameSemaphore // Release slot
		return
	}

	if err := game.AddPlayer(player2Req.Player); err != nil {
		gb.respondWithError(player1Req, player2Req, err)
		<-gb.gameSemaphore // Release slot
		return
	}

	// Start game
	if err := game.StartGame(); err != nil {
		gb.respondWithError(player1Req, player2Req, err)
		<-gb.gameSemaphore // Release slot
		return
	}

	// Create game session
	gameCtx, gameCancel := context.WithTimeout(gb.ctx, gb.gameTimeout)
	session := &GameSession{
		Game:      game,
		Context:   gameCtx,
		Cancel:    gameCancel,
		StartTime: time.Now(),
		Players:   []*Player{player1Req.Player, player2Req.Player},
	}

	// Register game session
	gb.gamesMutex.Lock()
	gb.activeGames[gameID] = session
	gb.gamesMutex.Unlock()

	// Start game management goroutine
	go gb.manageGameSession(session)

	// Respond to both players
	player1Req.Response <- &MatchmakingResponse{Game: game, Error: nil}
	player2Req.Response <- &MatchmakingResponse{
		Game:  game,
		Error: nil,
	}

	log.Printf("Created game %s between %s and %s",
		gameID, player1Req.Player.ID, player2Req.Player.ID)
}

// manageGameSession handles a single game's lifecycle
func (gb *GameBroker) manageGameSession(session *GameSession) {
	defer func() {
		// Cleanup when game ends
		gb.gamesMutex.Lock()
		delete(gb.activeGames, session.Game.ID)
		gb.gamesMutex.Unlock()

		session.Cancel()
		<-gb.gameSemaphore // Release slot

		log.Printf("Game %s ended after %v",
			session.Game.ID, time.Since(session.StartTime))
	}()

	// Monitor game state
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if game is finished
			if session.Game.State == GameStateFinished {
				log.Printf("Game %s finished, winner: %s",
					session.Game.ID, session.Game.Winner.ID)
				return
			}

		case <-session.Context.Done():
			// Game timeout or cancellation
			session.Game.State = GameStateFinished
			log.Printf("Game %s timed out or cancelled", session.Game.ID)
			return
		}
	}
}

// gameCleanupWorker periodically cleans up stale games
func (gb *GameBroker) gameCleanupWorker() {
	defer gb.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			gb.cleanupStaleGames()
		case <-gb.ctx.Done():
			return
		}
	}
}

// cleanupStaleGames removes games that have been inactive too long
func (gb *GameBroker) cleanupStaleGames() {
	gb.gamesMutex.Lock()
	defer gb.gamesMutex.Unlock()

	now := time.Now()
	for gameID, session := range gb.activeGames {
		if now.Sub(session.StartTime) > gb.gameTimeout {
			log.Printf("Cleaning up stale game %s", gameID)
			session.Cancel()
			delete(gb.activeGames, gameID)
		}
	}
}

// monitoringWorker provides metrics and monitoring
func (gb *GameBroker) monitoringWorker() {
	defer gb.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			gb.logMetrics()
		case <-gb.ctx.Done():
			return
		}
	}
}

// logMetrics logs current broker metrics
func (gb *GameBroker) logMetrics() {
	gb.gamesMutex.RLock()
	activeCount := len(gb.activeGames)
	gb.gamesMutex.RUnlock()

	queueSize := len(gb.queue)
	availableSlots := len(gb.gameSemaphore)

	log.Printf("Broker metrics - Active games: %d, Queue: %d, Available slots: %d",
		activeCount, queueSize, availableSlots)
}

// Helper methods

func (gb *GameBroker) respondWithError(player1Req, player2Req *MatchmakingRequest, err error) {
	player1Req.Response <- &MatchmakingResponse{
		Error: err,
		Game:  nil,
	}
	player2Req.Response <- &MatchmakingResponse{
		Error: err,
		Game:  nil,
	}
}

// GetActiveGameCount returns the number of active games
func (gb *GameBroker) GetActiveGameCount() int {
	gb.gamesMutex.RLock()
	defer gb.gamesMutex.RUnlock()
	return len(gb.activeGames)
}

// GetQueueSize returns the current queue size
func (gb *GameBroker) GetQueueSize() int {
	return len(gb.queue)
}

// GetAvailableSlots returns the number of available game slots
func (gb *GameBroker) GetAvailableSlots() int {
	return len(gb.gameSemaphore)
}

// GetGameSession returns a game session by ID
func (gb *GameBroker) GetGameSession(gameID string) (*GameSession, bool) {
	gb.gamesMutex.RLock()
	defer gb.gamesMutex.RUnlock()
	session, exists := gb.activeGames[gameID]
	return session, exists
}

// Example usage and integration
func ExampleUsage() {
	// Create broker with max 100 concurrent games
	broker := NewGameBroker(100)
	broker.Start()
	defer broker.Stop()

	// Simulate players joining
	player1 := NewPlayer("player1", "Alice")
	player2 := NewPlayer("player2", "Bob")

	// Request games (these would typically be called from HTTP handlers)
	go func() {
		game, err := broker.RequestGame(player1)
		if err != nil {
			log.Printf("Player1 error: %v", err)
			return
		}
		log.Printf("Player1 joined game %s", game.ID)
	}()

	go func() {
		game, err := broker.RequestGame(player2)
		if err != nil {
			log.Printf("Player2 error: %v", err)
			return
		}
		log.Printf("Player2 joined game %s", game.ID)
	}()

	// Wait a bit for demonstration
	time.Sleep(2 * time.Second)
}
