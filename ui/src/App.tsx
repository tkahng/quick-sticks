import { useEffect, useState } from "react";
import useWebSocket from "react-use-websocket";
import "./App.css";
import { ModeToggle } from "./components/mode-toggle";
import { Providers } from "./components/providers";
type Message<T> = {
  type: string;
  data: T;
};
type GameState = "ready" | "in-progress" | "waiting" | "finished";
type Game = {
  // 	CurrentTurn int       `json:"currentTurn"` // 0 for player1, 1 for player2
  currentTurn: number;
  state: GameState;
  // State       GameState `json:"state"`
  winner: string | null;
  // Winner      *Player   `json:"winner,omitempty"`
  createdAt: Date;
  // CreatedAt   time.Time `json:"createdAt"`
};
type GameMessage = Message<Game>;

type ErrorMessage = Message<string>;

type IncomingMessage = GameMessage | ErrorMessage;
function App() {
  const [count, setCount] = useState(0);
  const WS_URL = "/api/ws";
  const [stats, setStats] = useState<Game | null>(null);
  const { lastJsonMessage } = useWebSocket<IncomingMessage | null>(WS_URL);
  const handleMessage = (msg: IncomingMessage | null) => {
    if (!msg) return;
    if (msg.type === "error") return;
    if (msg.type !== "game_state") {
      const state = msg.data as Game;
      setStats(state);
    }
  };
  useEffect(() => {
    if (!lastJsonMessage) return;
    handleMessage(lastJsonMessage);
    console.log(lastJsonMessage);
  }, [lastJsonMessage]);

  const newLocal = stats?.state || "waiting";
  return (
    <Providers>
      <nav className="flex justify-between p-4">
        <a href="/">Sticks</a>
        <ModeToggle />
      </nav>
      <div className="bg-background p-4 pt-16">
        <h1>Vite + React</h1>
        <p>current status: {newLocal}</p>
        <div className="card">
          <button onClick={() => setCount((count) => count + 1)}>
            count is {count}
          </button>
          <p>
            Edit <code>src/App.tsx</code> and save to test HMR
          </p>
        </div>
        <p className="read-the-docs">
          Click on the Vite and React logos to learn more
        </p>
      </div>
    </Providers>
  );
}

export default App;
