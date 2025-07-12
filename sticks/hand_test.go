package sticks

import (
	"testing"
)

func TestHand_Attack(t *testing.T) {
	type fields struct {
		Fingers int
	}
	type args struct {
		opp *Hand
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "attack hand with 0 fingers",
			fields: fields{
				Fingers: 4,
			},
			args: args{
				opp: &Hand{
					fingers: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "attack hand with 5 fingers",
			fields: fields{
				Fingers: 4,
			},
			args: args{
				opp: &Hand{
					fingers: 5,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Hand{
				fingers: tt.fields.Fingers,
			}
			if err := h.Attack(tt.args.opp); (err != nil) != tt.wantErr {
				t.Errorf("Hand.Attack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHand_Take(t *testing.T) {
	type fields struct {
		Fingers int
	}
	type args struct {
		other  *Hand
		points int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "hand with 5 fingers takes 1 from other with 1",
			fields: fields{
				Fingers: 5,
			},
			args: args{
				other: &Hand{
					fingers: 1,
				},
				points: 1,
			},
			wantErr: true,
		},
		{
			name: "hand with 1 fingers takes 1 from other with 2",
			fields: fields{
				Fingers: 1,
			},
			args: args{
				other: &Hand{
					fingers: 2,
				},
				points: 1,
			},
			wantErr: false,
		},
		{
			name: "hand with 2 fingers takes 1 from other with 0",
			fields: fields{
				Fingers: 2,
			},
			args: args{
				other: &Hand{
					fingers: 0,
				},
				points: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Hand{
				fingers: tt.fields.Fingers,
			}
			if err := h.Take(tt.args.other, tt.args.points); (err != nil) != tt.wantErr {
				t.Errorf("Hand.Split() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
