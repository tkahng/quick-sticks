package sticks

import (
	"testing"
)

func TestModulo(t *testing.T) {
	t.Run("5 divide by 2", func(t *testing.T) {
		got := 7 / 2
		want := 3
		if got != want {
			t.Errorf("Wanted %d but got %d", want, got)
		}
	})
}

func TestGame_AddPlayer(t *testing.T) {
	type fields struct {
		ID        string
		Players   [2]*Player
		TurnIndex int
		Status    string
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
				ID:        "empty game",
				Players:   [2]*Player{},
				TurnIndex: 0,
				Status:    "waiting",
			},
			args: args{
				p: &Player{
					ID:    "player 1",
					Idx:   0,
					Hands: [2]*Hand{},
				},
			},
			wantErr: false,
		},
		{
			name: "add player to waiting game",
			fields: fields{
				ID: "waiting game",
				Players: [2]*Player{
					{
						ID:    "player 1",
						Idx:   0,
						Hands: [2]*Hand{},
					},
				},
				TurnIndex: 0,
				Status:    "waiting",
			},
			args: args{
				p: &Player{
					ID:    "player 1",
					Idx:   0,
					Hands: [2]*Hand{},
				},
			},
			wantErr: false,
		},
		{
			name: "add player to full game",
			fields: fields{
				ID: "waiting game",
				Players: [2]*Player{
					{
						ID:    "player 1",
						Idx:   0,
						Hands: [2]*Hand{},
					},
					{
						ID:    "player 2",
						Idx:   1,
						Hands: [2]*Hand{},
					},
				},
				TurnIndex: 0,
				Status:    "waiting",
			},
			args: args{
				p: &Player{
					ID:    "player 1",
					Idx:   0,
					Hands: [2]*Hand{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{
				ID:        tt.fields.ID,
				Players:   tt.fields.Players,
				TurnIndex: tt.fields.TurnIndex,
				Status:    tt.fields.Status,
			}
			if err := g.AddPlayer(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Game.AddPlayer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
