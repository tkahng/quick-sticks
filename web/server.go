package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tkahng/sticks/sticks"
)

type GameActionType string

const (
	GameActionTypeAttack GameActionType = "attack"
	GameActionTypeSplit  GameActionType = "split"
)

// case 'game_start':
//     gameState = message.data;
//     document.getElementById('status').textContent = 'Game started!';
//     updateGameDisplay();
//     break;
// case 'game_state':
//     gameState = message.data;
//     updateGameDisplay();
//     break;
// case 'error':
//     alert('Error: ' + message.data);
//     break;
// case 'game_end':
//     gameState = message.data;
//     updateGameDisplay();
//     if (gameState.winner) {
//         const isWinner = (myPlayerNumber === 1 && gameState.winner.id === gameState.player1.id) ||
//                        (myPlayerNumber === 2 && gameState.winner.id === gameState.player2.id);
//         document.getElementById('status').textContent = isWinner ? 'You Win!' : 'You Lose!';
//     }
//     break;

type (
	GameActionInput struct {
		Type GameActionType `json:"type"`
		Data any            `json:"data"`
	}

	GameAttackData struct {
		WithLeft   bool `json:"with_left"`
		AttackLeft bool `json:"attack_left"`
		Points     int  `json:"points"`
	}
	GameSplitData struct {
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
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
			HandshakeTimeout: 0,
			ReadBufferSize:   0,
			WriteBufferSize:  0,
			WriteBufferPool:  nil,
			Subprotocols:     []string{},
			Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
				panic("TODO")
			},
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
	gs.mux.HandleFunc("/", gs.handleHome)
	gs.mux.HandleFunc("/ws", gs.handleWebSocket)
	gs.mux.HandleFunc("/api/stats", gs.handleStats)
	gs.mux.HandleFunc("/api/health", gs.handleHealth)
}

// handleHome serves the game client
func (gs *GameServer) handleHome(w http.ResponseWriter, r *http.Request) {
	// Serve your HTML game client here
	w.Header().Set("Content-Type", "text/html")
	// nolint:errcheck
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head><title>Chopsticks Game</title></head>
		<body>
			<h1>Chopsticks Game</h1>
			<div id="status">Connecting...</div>
			<script>
				const ws = new WebSocket('ws://localhost:8080/ws');
				const status = document.getElementById('status');
				
				ws.onopen = function() {
					status.textContent = 'Connected - Finding opponent...';
				};
				
				ws.onmessage = function(event) {
					const msg = JSON.parse(event.data);
					console.log('Received:', msg);
					
					if (msg.type === 'game_matched') {
						status.textContent = 'Game found! Starting...';
					} else if (msg.type === 'game_state') {
						status.textContent = 'Game in progress';
					} else if (msg.type === 'error') {
						status.textContent = 'Error: ' + msg.data;
					}
				};
				
				ws.onclose = function() {
					status.textContent = 'Disconnected';
				};
			</script>
		</body>
		</html>
	`))
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
	playerID := generatePlayerID()
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
		var msg GameActionInput
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
func (gs *GameServer) processGameAction(game *sticks.Game, player *sticks.Player, msg GameActionInput) error {
	actionType := msg.Type

	// Verify it's the player's turn
	currentPlayer := game.GetCurrentPlayer()
	if currentPlayer.ID != player.ID {
		return fmt.Errorf("not your turn")
	}

	switch actionType {
	case GameActionTypeAttack:
		var data GameAttackData
		var ok bool
		if data, ok = msg.Data.(GameAttackData); !ok {
			return fmt.Errorf("invalid action data")
		}
		attackerIsLeft := data.WithLeft
		defenderIsLeft := data.AttackLeft

		return game.Attack(attackerIsLeft, defenderIsLeft)

	case GameActionTypeSplit:
		var data GameSplitData
		var ok bool
		if data, ok = msg.Data.(GameSplitData); !ok {
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
