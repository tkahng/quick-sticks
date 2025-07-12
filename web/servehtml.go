package web

import (
	"fmt"
	"net/http"
)

// Serve the static HTML file
func ServeHTML(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Chopsticks Game</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .game-board { max-width: 600px; margin: 0 auto; }
        .opponent-hands, .player-hands { display: flex; justify-content: center; margin: 20px 0; }
        .hand { margin: 0 20px; padding: 20px; border: 2px solid #333; border-radius: 10px; min-width: 80px; text-align: center; }
        .hand.alive { background-color: #90EE90; cursor: pointer; }
        .hand.dead { background-color: #FFB6C1; opacity: 0.5; }
        .hand.attacking { background-color: #FFD700; }
        .game-info { text-align: center; margin: 20px 0; }
        .controls { text-align: center; margin: 20px 0; }
        button { padding: 10px 20px; margin: 5px; font-size: 16px; }
        .status { padding: 10px; margin: 10px 0; background-color: #f0f0f0; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="game-board">
        <h1>Chopsticks Game</h1>
        <div class="status" id="status">Connecting...</div>
        
        <div class="opponent-hands">
            <div class="hand dead" id="opp-left">
                <div>Left Hand</div>
                <div id="opp-left-points">0</div>
            </div>
            <div class="hand dead" id="opp-right">
                <div>Right Hand</div>
                <div id="opp-right-points">0</div>
            </div>
        </div>
        
        <div class="game-info">
            <div id="turn-info">Waiting for game...</div>
        </div>
        
        <div class="player-hands">
            <div class="hand dead" id="player-left">
                <div>Left Hand</div>
                <div id="player-left-points">0</div>
            </div>
            <div class="hand dead" id="player-right">
                <div>Right Hand</div>
                <div id="player-right-points">0</div>
            </div>
        </div>
        
        <div class="controls">
            <button onclick="showSplitModal()">Split</button>
            <button onclick="resetSelection()">Reset Selection</button>
        </div>
    </div>

    <script>
        const ws = new WebSocket('ws://localhost:8080/ws');
        let gameState = null;
        let selectedHand = null;
        let isMyTurn = false;
        let myPlayerNumber = null;

        ws.onopen = function() {
            document.getElementById('status').textContent = 'Connected - Waiting for opponent...';
        };

        ws.onmessage = function(event) {
            const message = JSON.parse(event.data);
            handleMessage(message);
        };

        ws.onclose = function() {
            document.getElementById('status').textContent = 'Disconnected';
        };

        function handleMessage(message) {
            switch(message.type) {
                case 'game_start':
                    gameState = message.data;
                    document.getElementById('status').textContent = 'Game started!';
                    updateGameDisplay();
                    break;
                case 'game_state':
                    gameState = message.data;
                    updateGameDisplay();
                    break;
                case 'error':
                    alert('Error: ' + message.data);
                    break;
                case 'game_end':
                    gameState = message.data;
                    updateGameDisplay();
                    if (gameState.winner) {
                        const isWinner = (myPlayerNumber === 1 && gameState.winner.id === gameState.player1.id) ||
                                       (myPlayerNumber === 2 && gameState.winner.id === gameState.player2.id);
                        document.getElementById('status').textContent = isWinner ? 'You Win!' : 'You Lose!';
                    }
                    break;
            }
        }

        function updateGameDisplay() {
            if (!gameState) return;

            // Determine which player is me
            // For simplicity, assume player1 is always "me" in this demo
            const me = gameState.player1;
            const opponent = gameState.player2;
            
            myPlayerNumber = 1;
            isMyTurn = gameState.currentTurn === 0;


            updateHand('opp-left', opponent.leftHand);
            updateHand('opp-right', opponent.rightHand);


            updateHand('player-left', me.leftHand);
            updateHand('player-right', me.rightHand);

    
            document.getElementById('turn-info').textContent = 
                isMyTurn ? 'Your Turn' : 'Opponent\'s Turn';
        }

        function updateHand(elementId, hand) {
            const element = document.getElementById(elementId);
            const pointsElement = document.getElementById(elementId + '-points');
            
            pointsElement.textContent = hand.points;
            
            if (hand.alive) {
                element.className = 'hand alive';
            } else {
                element.className = 'hand dead';
            }
        }

        // Add click handlers for hands
        document.getElementById('player-left').onclick = function() {
            if (isMyTurn && gameState.player1.leftHand.alive) {
                selectHand('player-left', true);
            }
        };

        document.getElementById('player-right').onclick = function() {
            if (isMyTurn && gameState.player1.rightHand.alive) {
                selectHand('player-right', false);
            }
        };

        document.getElementById('opp-left').onclick = function() {
            console.log(isMyTurn, selectedHand, gameState.player2.leftHand.alive);
            if (isMyTurn && selectedHand && gameState.player2.leftHand.alive) {
                attack(selectedHand.isLeft, true);
            }
        };

        document.getElementById('opp-right').onclick = function() {
            if (isMyTurn && selectedHand && gameState.player2.rightHand.alive) {
                attack(selectedHand.isLeft, false);
            }
        };

        function selectHand(elementId, isLeft) {
            resetSelection();
            selectedHand = { elementId, isLeft };
            document.getElementById(elementId).classList.add('attacking');
        }

        function resetSelection() {
            if (selectedHand) {
                document.getElementById(selectedHand.elementId).classList.remove('attacking');
                selectedHand = null;
            }
        }

        function attack(attackerIsLeft, defenderIsLeft) {
            if (!isMyTurn || !selectedHand) return;

            const message = {
                type: 'attack',
                data: {
                    attackerIsLeft: attackerIsLeft,
                    defenderIsLeft: defenderIsLeft
                }
            };

            ws.send(JSON.stringify(message));
            resetSelection();
        }

        function showSplitModal() {
            if (!isMyTurn) return;

            const leftPoints = gameState.player1.leftHand.points;
            const rightPoints = gameState.player1.rightHand.points;
            const total = leftPoints + rightPoints;

            const newLeft = prompt('Enter new left hand points (current: ' + leftPoints + ', total: ' + total + '):');
            if (newLeft === null) return;

            const newRight = total - parseInt(newLeft);
            
            if (newLeft < 0 || newRight < 0 || newLeft + newRight !== total) {
                alert('Invalid split!');
                return;
            }

            const message = {
                type: 'split',
                data: {
                    fromLeft: true,
                    newLeftPoints: parseInt(newLeft),
                    newRightPoints: newRight
                }
            };

            ws.send(JSON.stringify(message));
        }
    </script>
</body>
</html>

`
	w.Header().Set("Content-Type", "text/html")
	count, err := w.Write([]byte(html))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Sent %d bytes of HTML\n", count)
}
