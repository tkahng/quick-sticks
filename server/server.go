package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tkahng/sticks"
)

type MessageType string

const (
	MessageTypeAttack         MessageType = "attack"
	MessageTypeSplit          MessageType = "split"
	MessageTypeStateGameStart MessageType = "state_game_start"
	MessageTypeGameState      MessageType = "state_game_state"
	MessageTypeError          MessageType = "error"
	MessageTypeGameEnd        MessageType = "game_end"
)

type (
	Message struct {
		Type MessageType `json:"type"`
		Data any         `json:"data"`
	}
	AttackMessageData struct {
		WithLeft   bool `json:"with_left"`
		AttackLeft bool `json:"attack_left"`
	}
	SplitMessageData struct {
		WithLeft bool `json:"with_left"`
		Points   int  `json:"points"`
	}
)

// GameServer integrates the matchmaking system with HTTP/WebSocket
type GameServer struct {
	broker   *sticks.GameBroker
	upgrader websocket.Upgrader
	mux      *http.ServeMux
}

func (gs *GameServer) Hanlder() http.Handler {
	return gs.mux
}

// NewGameServer creates a new game server
func NewGameServer(maxConcurrentGames int) *GameServer {
	broker := sticks.NewGameBroker(maxConcurrentGames)

	return &GameServer{
		broker: broker,
		upgrader: websocket.Upgrader{
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			HandshakeTimeout:  0,
			WriteBufferPool:   nil,
			Subprotocols:      nil,
			Error:             nil,
			CheckOrigin:       nil,
			EnableCompression: false,
		},
		mux: http.NewServeMux(),
	}
}

// Start starts the game server
func (gs *GameServer) Start() {
	gs.broker.Start()
	gs.setupRoutes()
}

// Stop gracefully stops the game server
func (gs *GameServer) Stop() {
	gs.broker.Stop()
}

// setupRoutes configures HTTP routes
func (gs *GameServer) setupRoutes() {
	// gs.mux.HandleFunc("/", gs.handleHome)
	gs.mux.HandleFunc("/api/ws", PlayerID(http.HandlerFunc(gs.handleWebSocket)).ServeHTTP)
	gs.mux.HandleFunc("/api/stats", gs.handleStats)
	gs.mux.HandleFunc("/api/health", gs.handleHealth)
}

// handleWebSocket handles WebSocket connections for real-time gameplay
func (gs *GameServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := gs.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	// nolint:errcheck
	defer conn.Close()

	// Create player
	playerID := getPlayerIDFromContext(r.Context())
	if playerID == "" {
		gs.sendError(conn, "Player ID not found")
		return
	}
	player := sticks.NewPlayer(playerID, "Player")

	// Set up WebSocket connection for the player
	// (You'll need to add a SetConnection method to your Player struct)

	log.Printf("Player %s connected", playerID)

	// Request game from matchmaking
	gameReady := make(chan *sticks.Game, 1)
	go func() {
		game, err := gs.broker.RequestGame(player)
		if err != nil {
			log.Printf("Matchmaking error for player %s: %v", playerID, err)
			gs.sendError(conn, err.Error())
			return
		}
		gameReady <- game
	}()

	// Handle the game lifecycle
	select {
	case game := <-gameReady:
		gs.handleGameSession(conn, player, game)
	case <-time.After(30 * time.Second):
		gs.sendError(conn, "Matchmaking timeout")
		return
	}
}

// handleGameSession manages a player's game session
func (gs *GameServer) handleGameSession(conn *websocket.Conn, player *sticks.Player, game *sticks.Game) {
	// Notify player that game was found
	gs.sendMessage(conn, "game_matched", map[string]any{
		"gameId": game.ID,
		"player": player,
	})

	// Send initial game state
	gs.sendGameState(conn, game)

	// Handle game messages
	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Process game actions
		if err := gs.processGameAction(game, player, msg); err != nil {
			gs.sendError(conn, err.Error())
			continue
		}

		// Send updated game state
		gs.sendGameState(conn, game)

		// Check if game is finished
		if game.State == sticks.GameStateFinished {
			gs.sendMessage(conn, "game_end", map[string]any{
				"winner": game.Winner,
			})
			break
		}
	}
}

// processGameAction processes a game action from a player
func (gs *GameServer) processGameAction(game *sticks.Game, player *sticks.Player, msg Message) error {
	actionType := msg.Type

	// Verify it's the player's turn
	currentPlayer := game.GetCurrentPlayer()
	if currentPlayer.ID != player.ID {
		return fmt.Errorf("not your turn")
	}

	switch actionType {
	case MessageTypeAttack:
		var data AttackMessageData
		var ok bool
		if data, ok = msg.Data.(AttackMessageData); !ok {
			return fmt.Errorf("invalid action data")
		}
		attackerIsLeft := data.WithLeft
		defenderIsLeft := data.AttackLeft

		return game.Attack(attackerIsLeft, defenderIsLeft)

	case MessageTypeSplit:
		var data SplitMessageData
		var ok bool
		if data, ok = msg.Data.(SplitMessageData); !ok {
			return fmt.Errorf("invalid action data")
		}
		fromLeft := data.WithLeft
		points := data.Points

		return game.Split(fromLeft, points)

	default:
		return fmt.Errorf("unknown action type: %s", actionType)
	}
}

// handleStats provides server statistics
func (gs *GameServer) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]any{
		"activeGames":    gs.broker.GetActiveGameCount(),
		"queueSize":      gs.broker.GetQueueSize(),
		"availableSlots": gs.broker.GetAvailableSlots(),
		"timestamp":      time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	// nolint:errcheck
	json.NewEncoder(w).Encode(stats)
}

// Global variables for tracking
var startTime = time.Now()

// handleHealth provides health check endpoint
func (gs *GameServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]any{
		"status": "ok",
		"uptime": time.Since(startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	// nolint:errcheck
	json.NewEncoder(w).Encode(health)
}

// Helper methods

func (gs *GameServer) sendMessage(conn *websocket.Conn, msgType string, data any) {
	msg := map[string]interface{}{
		"type": msgType,
		"data": data,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (gs *GameServer) sendError(conn *websocket.Conn, errorMsg string) {
	gs.sendMessage(conn, "error", errorMsg)
}

func (gs *GameServer) sendGameState(conn *websocket.Conn, game *sticks.Game) {
	gs.sendMessage(conn, "game_state", game)
}

func generatePlayerID() string {
	return fmt.Sprintf("player_%d", time.Now().UnixNano())
}
