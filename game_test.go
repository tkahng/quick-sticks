package sticks

import (
	"errors"
	"testing"
)

func TestGame_AddPlayer(t *testing.T) {
	type fields struct {
		ID      string
		Player1 *Player
		Player2 *Player
	}
	type args struct {
		p *Player
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "add player to empty game",
			fields: fields{
				ID:      "empty game",
				Player1: nil,
				Player2: nil,
			},
			args: args{
				p: NewPlayer("player 1", ""),
			},
			wantErr: false,
		},
		{
			name: "add player to waiting game",
			fields: fields{
				ID:      "waiting game",
				Player1: NewPlayer("player 1", ""),
				Player2: nil,
			},
			args: args{
				p: NewPlayer("player 2", ""),
			},
			wantErr: false,
		},
		{
			name: "add player to full game",
			fields: fields{
				ID:      "waiting game",
				Player1: NewPlayer("player 1", ""),
				Player2: NewPlayer("player 2", ""),
			},
			args: args{
				p: NewPlayer("player 3", ""),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame(tt.fields.ID)
			if tt.fields.Player1 != nil {
				err := g.AddPlayer(tt.fields.Player1)
				if err != nil {
					t.Errorf("Game.AddPlayer() error = %v", err)
				}
			}
			if tt.fields.Player2 != nil {
				err := g.AddPlayer(tt.fields.Player2)
				if err != nil {
					t.Errorf("Game.AddPlayer() error = %v", err)
				}
			}
			if err := g.AddPlayer(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Game.AddPlayer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGame_Attack(t *testing.T) {
	player1 := NewPlayer("player 1", "")
	player2 := NewPlayer("player 2", "")
	game := NewGame("game")

	if err := errors.Join(game.AddPlayer(player1), game.AddPlayer(player2)); err != nil {
		t.Errorf("Game.AddPlayer() error = %v", err)
	}

	if err := game.StartGame(); err != nil {
		t.Errorf("Game.StartGame() error = %v", err)
	}

	// 1,1
	// 2,1
	err := game.Attack(true, true)
	if err != nil {
		t.Errorf("Game.Attack() error = %v", err)
	}
	game.PrintScore()

	// 2,1
	// 3,1
	err = game.Attack(true, true)
	if err != nil {
		t.Errorf("Game.Attack() error = %v", err)
	}
	game.PrintScore()
	// 3,1
	// 5,1
	err = game.Attack(true, true)
	if err != nil {
		t.Errorf("Game.Attack() error = %v", err)
	}
	// game.
	game.PrintScore()
}
